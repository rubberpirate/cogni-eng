package delivery

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"learning.local/workspace/internal/access"
)

type SMTP struct {
	Address, From, Username, Password string
	Local                             bool
}

// Send reports SMTP acceptance; it does not claim inbox delivery.
func (s SMTP) Send(ctx context.Context, id, to, subject, body string) error {
	if s.Address == "" || strings.ContainsAny(s.From+to+subject, "\r\n") {
		return errors.New("invalid SMTP configuration or message")
	}
	host, _, err := net.SplitHostPort(s.Address)
	if err != nil {
		return err
	}
	d := net.Dialer{Timeout: 10 * time.Second}
	conn, err := d.DialContext(ctx, "tcp", s.Address)
	if err != nil {
		return err
	}
	defer conn.Close()
	if err = conn.SetDeadline(time.Now().Add(15 * time.Second)); err != nil {
		return err
	}
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()
	if !s.Local {
		if err = client.StartTLS(&tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}); err != nil {
			return err
		}
	}
	if s.Username != "" {
		if err = client.Auth(smtp.PlainAuth("", s.Username, s.Password, host)); err != nil {
			return err
		}
	}
	if err = client.Mail(s.From); err != nil {
		return err
	}
	if err = client.Rcpt(to); err != nil {
		return err
	}
	out, err := client.Data()
	if err != nil {
		return err
	}
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMessage-ID: <%s@learning-workspace.invalid>\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s", s.From, to, subject, id, body)
	if _, err = out.Write([]byte(msg)); err != nil {
		return err
	}
	if err = out.Close(); err != nil {
		return err
	}
	return client.Quit()
}

type Worker struct {
	DB     *pgxpool.Pool
	Sender SMTP
}

func (w Worker) Run(ctx context.Context) error {
	timer := time.NewTicker(time.Second)
	defer timer.Stop()
	for {
		if err := w.Once(ctx); err != nil && ctx.Err() == nil {
			return err
		}
		select {
		case <-ctx.Done():
			return nil
		case <-timer.C:
		}
	}
}
func (w Worker) Once(ctx context.Context) error {
	var id, recipient, kind, lease string
	var payload []byte
	var attempts int
	lease = access.Random()
	err := w.DB.QueryRow(ctx, `UPDATE delivery_job SET state='running',attempts=attempts+1,lease_until=now()+interval '60 seconds',lease_token=$1 WHERE id=(SELECT id FROM delivery_job WHERE (state='queued' AND run_at<=now()) OR (state='running' AND lease_until<now()) ORDER BY run_at FOR UPDATE SKIP LOCKED LIMIT 1) RETURNING id,recipient_id,kind,payload,attempts`, lease).Scan(&id, &recipient, &kind, &payload, &attempts)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	var email, state string
	err = w.DB.QueryRow(ctx, "SELECT email,state FROM person WHERE id=$1", recipient).Scan(&email, &state)
	if err != nil {
		return err
	}
	var data struct {
		Token  string `json:"token"`
		Origin string `json:"origin"`
		To     string `json:"to"`
	}
	if err = json.Unmarshal(payload, &data); err != nil {
		return w.finish(ctx, id, lease, attempts, errors.New("invalid payload"))
	}
	if state == "suspended" || state == "archived" {
		_, err = w.DB.Exec(ctx, "UPDATE delivery_job SET state='cancelled',payload='{}',last_error='Account unavailable' WHERE id=$1 AND lease_token=$2", id, lease)
		return err
	}
	page := "reset"
	subject := "Reset your Learning Workspace password"
	if kind == "invitation" {
		page = "join"
		subject = "Your Learning Workspace invitation"
	}
	if kind == "email_change" {
		page = "confirm-email"
		subject = "Confirm your new Learning Workspace email"
		email = data.To
	}
	prefix := ""
	if w.Sender.Local {
		prefix = "DEVELOPMENT MAIL CATCHER — no live delivery\n\n"
	}
	err = w.Sender.Send(ctx, id, email, subject, prefix+data.Origin+"/"+page+"#"+data.Token+"\n\nThis link is single-use. If unexpected, contact your institution.")
	return w.finish(ctx, id, lease, attempts, err)
}
func (w Worker) finish(ctx context.Context, id, lease string, attempts int, sendErr error) error {
	if sendErr == nil {
		_, err := w.DB.Exec(ctx, "UPDATE delivery_job SET state='accepted',payload='{}',lease_until=NULL,last_error='' WHERE id=$1 AND lease_token=$2 AND state='running'", id, lease)
		return err
	}
	state := "queued"
	if attempts >= 5 {
		state = "failed"
	}
	delay := time.Duration(1<<min(attempts, 8)) * time.Second
	// Keep provider details out of ordinary operations responses.
	_, err := w.DB.Exec(ctx, "UPDATE delivery_job SET state=$3,run_at=$4,lease_until=NULL,last_error='SMTP acceptance failed; check provider configuration and connectivity' WHERE id=$1 AND lease_token=$2 AND state='running'", id, lease, state, time.Now().Add(delay))
	return err
}
