package adapters

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// MySQLAdapter implements DatabaseAdapter for MySQL using mysqldump and mysql client.
type MySQLAdapter struct {
	Host           string
	Port           int
	User           string
	Password       string
	DBName         string
	MySQLDumpPath  string
	MySQLPath      string
	MySQLAdminPath string
}

// NewMySQLAdapter creates a new MySQLAdapter instance.
func NewMySQLAdapter(host string, port int, user, password, dbName string) *MySQLAdapter {
	if port <= 0 {
		port = 3306
	}
	return &MySQLAdapter{
		Host:           host,
		Port:           port,
		User:           user,
		Password:       password,
		DBName:         dbName,
		MySQLDumpPath:  "mysqldump",
		MySQLPath:      "mysql",
		MySQLAdminPath: "mysqladmin",
	}
}

// TestConnection tests whether the MySQL server is reachable and accepting connections.
func (m *MySQLAdapter) TestConnection() error {
	addr := net.JoinHostPort(m.Host, strconv.Itoa(m.Port))
	// 1. Basic TCP connectivity check
	conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
	if err != nil {
		return fmt.Errorf("unable to reach MySQL server at %s: %w", addr, err)
	}
	_ = conn.Close()

	// 2. Ping using mysqladmin if available
	if _, lookErr := exec.LookPath(m.MySQLAdminPath); lookErr == nil {
		cmd := exec.Command(m.MySQLAdminPath,
			"-h", m.Host,
			"-P", strconv.Itoa(m.Port),
			"-u", m.User,
			"ping",
		)
		if m.Password != "" {
			cmd.Env = append(os.Environ(), fmt.Sprintf("MYSQL_PWD=%s", m.Password))
		}
		out, runErr := cmd.CombinedOutput()
		if runErr != nil {
			return fmt.Errorf("mysqladmin ping failed: %w (output: %s)", runErr, strings.TrimSpace(string(out)))
		}
	}

	return nil
}

// Backup executes mysqldump and streams the output directly.
// Uses --single-transaction for InnoDB consistency without global table locks,
// and --quick to stream rows without buffering entire tables in memory.
func (m *MySQLAdapter) Backup(backupType string, tables []string) (io.ReadCloser, error) {
	args := []string{
		"-h", m.Host,
		"-P", strconv.Itoa(m.Port),
		"-u", m.User,
		"--single-transaction",
		"--quick",
		"--routines",
		"--triggers",
		m.DBName,
	}

	for _, tbl := range tables {
		tbl = strings.TrimSpace(tbl)
		if tbl != "" {
			args = append(args, tbl)
		}
	}

	cmd := exec.Command(m.MySQLDumpPath, args...)

	// Pass password via MYSQL_PWD environment variable to prevent CLI exposure
	if m.Password != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("MYSQL_PWD=%s", m.Password))
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe for mysqldump: %w", err)
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start mysqldump: %w", err)
	}

	return &execReadCloser{
		Reader: stdout,
		cmd:    cmd,
		stderr: &stderr,
	}, nil
}

// Restore executes the mysql client, reading the SQL stream from stdin.
func (m *MySQLAdapter) Restore(stream io.Reader, tables []string) error {
	args := []string{
		"-h", m.Host,
		"-P", strconv.Itoa(m.Port),
		"-u", m.User,
		m.DBName,
	}

	cmd := exec.Command(m.MySQLPath, args...)
	if m.Password != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("MYSQL_PWD=%s", m.Password))
	}
	cmd.Stdin = stream

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg != "" {
			return fmt.Errorf("mysql restore failed: %w: %s", err, errMsg)
		}
		return fmt.Errorf("mysql restore failed: %w", err)
	}

	return nil
}

// ListTables queries the MySQL database to list all tables.
func (m *MySQLAdapter) ListTables() ([]string, error) {
	args := []string{
		"-h", m.Host,
		"-P", strconv.Itoa(m.Port),
		"-u", m.User,
		"-s",
		"-N",
		"-e", "SHOW TABLES;",
		m.DBName,
	}

	cmd := exec.Command(m.MySQLPath, args...)
	if m.Password != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("MYSQL_PWD=%s", m.Password))
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("listing MySQL tables failed: %w: %s", err, strings.TrimSpace(stderr.String()))
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
