package adapters

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// PostgresAdapter implements DatabaseAdapter for PostgreSQL using pg_dump and pg_restore.
type PostgresAdapter struct {
	Host          string
	Port          int
	User          string
	Password      string
	DBName        string
	PgDumpPath    string
	PgRestorePath string
	PgIsReadyPath string
	PsqlPath      string
}

// resolvePostgresBinary finds the executable from PATH or standard Windows installations.
func resolvePostgresBinary(binaryName string) string {
	if path, err := exec.LookPath(binaryName); err == nil {
		return path
	}

	// Common Windows installation locations
	candidates := []string{
		`C:\Program Files\PostgreSQL\17\bin`,
		`C:\Program Files\PostgreSQL\16\bin`,
		`C:\Program Files\PostgreSQL\15\bin`,
	}

	for _, dir := range candidates {
		target := filepath.Join(dir, binaryName+".exe")
		if _, err := os.Stat(target); err == nil {
			return target
		}
	}

	return binaryName
}

// NewPostgresAdapter creates a new PostgresAdapter instance with resolved binary paths.
func NewPostgresAdapter(host string, port int, user, password, dbName string) *PostgresAdapter {
	if port <= 0 {
		port = 5432
	}
	return &PostgresAdapter{
		Host:          host,
		Port:          port,
		User:          user,
		Password:      password,
		DBName:        dbName,
		PgDumpPath:    resolvePostgresBinary("pg_dump"),
		PgRestorePath: resolvePostgresBinary("pg_restore"),
		PgIsReadyPath: resolvePostgresBinary("pg_isready"),
		PsqlPath:      resolvePostgresBinary("psql"),
	}
}

// TestConnection tests whether PostgreSQL is reachable and accepting connections.
func (p *PostgresAdapter) TestConnection() error {
	addr := net.JoinHostPort(p.Host, strconv.Itoa(p.Port))
	// 1. Basic TCP connectivity check
	conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
	if err != nil {
		return fmt.Errorf("unable to reach postgres host at %s: %w", addr, err)
	}
	_ = conn.Close()

	// 2. If pg_isready is found, check if server is accepting connections
	cmd := exec.Command(p.PgIsReadyPath,
		"-h", p.Host,
		"-p", strconv.Itoa(p.Port),
		"-d", p.DBName,
		"-U", p.User,
	)
	out, runErr := cmd.CombinedOutput()
	if runErr != nil {
		return fmt.Errorf("pg_isready check failed: %w (output: %s)", runErr, strings.TrimSpace(string(out)))
	}

	return nil
}

// Backup executes pg_dump and streams the raw dump output.
// Uses custom format (-Fc) by default to support selective table restores, compression, and TOC reordering.
func (p *PostgresAdapter) Backup(backupType string, tables []string) (io.ReadCloser, error) {
	args := []string{
		"-h", p.Host,
		"-p", strconv.Itoa(p.Port),
		"-U", p.User,
		"-d", p.DBName,
	}

	if strings.ToLower(backupType) == "sql" || strings.ToLower(backupType) == "plain" {
		args = append(args, "-Fp")
	} else {
		// Custom format (-Fc) is the default: supports pg_restore selective tables & compression
		args = append(args, "-Fc")
	}

	for _, tbl := range tables {
		tbl = strings.TrimSpace(tbl)
		if tbl != "" {
			args = append(args, "-t", tbl)
		}
	}

	cmd := exec.Command(p.PgDumpPath, args...)

	// Pass password via environment variable to avoid process listing exposure
	if p.Password != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", p.Password))
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe for pg_dump: %w", err)
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start pg_dump: %w", err)
	}

	return &execReadCloser{
		Reader: stdout,
		cmd:    cmd,
		stderr: &stderr,
	}, nil
}

// Restore executes pg_restore reading the dump stream from stdin.
func (p *PostgresAdapter) Restore(stream io.Reader, tables []string) error {
	// 1. Ensure target database exists in case the entire database was accidentally dropped.
	// Connect to default maintenance database 'postgres' to issue CREATE DATABASE if needed.
	createDBCmd := exec.Command(p.PsqlPath,
		"-h", p.Host,
		"-p", strconv.Itoa(p.Port),
		"-U", p.User,
		"-d", "postgres",
		"-c", fmt.Sprintf("CREATE DATABASE \"%s\";", p.DBName),
	)
	if p.Password != "" {
		createDBCmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", p.Password))
	}
	// Ignore error if database already exists
	_ = createDBCmd.Run()

	args := []string{
		"-h", p.Host,
		"-p", strconv.Itoa(p.Port),
		"-U", p.User,
		"-d", p.DBName,
		"--clean",
		"--if-exists",
	}

	for _, tbl := range tables {
		tbl = strings.TrimSpace(tbl)
		if tbl != "" {
			args = append(args, "-t", tbl)
		}
	}

	cmd := exec.Command(p.PgRestorePath, args...)
	if p.Password != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", p.Password))
	}
	cmd.Stdin = stream

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg != "" {
			return fmt.Errorf("pg_restore failed: %w: %s", err, errMsg)
		}
		return fmt.Errorf("pg_restore failed: %w", err)
	}

	return nil
}

// ListTables queries the database to list all public user tables.
func (p *PostgresAdapter) ListTables() ([]string, error) {
	query := "SELECT tablename FROM pg_tables WHERE schemaname = 'public' ORDER BY tablename;"
	cmd := exec.Command(p.PsqlPath,
		"-h", p.Host,
		"-p", strconv.Itoa(p.Port),
		"-U", p.User,
		"-d", p.DBName,
		"-t", "-A", "-c", query,
	)
	if p.Password != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", p.Password))
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("listing tables failed: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	var tables []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			tables = append(tables, l)
		}
	}
	return tables, nil
}

// execReadCloser wraps stdout reader of an exec.Cmd and waits for process completion on Close.
type execReadCloser struct {
	io.Reader
	cmd    *exec.Cmd
	stderr *bytes.Buffer
}

func (e *execReadCloser) Close() error {
	err := e.cmd.Wait()
	if err != nil {
		errMsg := strings.TrimSpace(e.stderr.String())
		if errMsg != "" {
			return fmt.Errorf("process failed (%w): %s", err, errMsg)
		}
		return fmt.Errorf("process failed: %w", err)
	}
	return nil
}
