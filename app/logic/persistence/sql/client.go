package sql

import (
	"context"
	"mercury/app/logic/persistence"
	"mercury/x"
	"mercury/x/database/sqlx"
	"mercury/x/ecode"
	"strings"
	"time"
)

type clientPersister struct {
	db *sqlx.DB
}

const (
	insertClientSQL = `
INSERT INTO
    client (
		id,
        created_at,
        updated_at,
        name,
		token_secret,
		token_expire,
        credential,
		user_count
    )
VALUES
    ($1, $2, $2, $3, $4, $5, $6, 0);
`
)

func increaseClientUserCount(tx *sqlx.Tx, id string, count int64) error {
	return tx.Exec("UPDATE client SET user_count = user_count + $1 WHERE id = $2;", 1, count, id)
}

func increaseClientGroupCount(tx *sqlx.Tx, id string, count int64) error {
	return tx.Exec("UPDATE client SET group_count = group_count + $1 WHERE id = $2;", 1, count, id)
}

func (p *clientPersister) GetClientCredential(_ context.Context, id string) (string, error) {
	var credential string
	if err := p.db.QueryRow("SELECT credential FROM client WHERE id = $1;", id).Scan(&credential); sqlx.IsErrNoRows(err) {
		return "", ecode.ErrDataDoesNotExist
	} else if err != nil {
		return "", err
	}

	return credential, nil
}

func (p *clientPersister) GetClient(_ context.Context, id string) (*persistence.Client, error) {
	var (
		name, tokenSecret                                        string
		createdAt, updatedAt, tokenExpire, userCount, groupCount int64
	)
	if err := p.db.QueryRow("SELECT created_at, updated_at, name, token_expire, token_secret, user_count, group_count FROM client WHERE id = $1;", id).
		Scan(&createdAt, &updatedAt, &name, &tokenExpire, &tokenSecret, &userCount, &groupCount); sqlx.IsErrNoRows(err) {
		return nil, ecode.ErrDataDoesNotExist
	} else if err != nil {
		return nil, err
	}

	return &persistence.Client{
		ID:          id,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		Name:        name,
		TokenSecret: []byte(tokenSecret),
		TokenExpire: time.Duration(tokenExpire) * time.Second,
		UserCount:   userCount,
		GroupCount:  groupCount,
	}, nil
}

func (p *clientPersister) Create(_ context.Context, in *persistence.ClientCreate) error {
	var isExist int
	if err := p.db.QueryRow("SELECT 1 FROM client WHERE name = $1 limit 1;", in.Name).
		Scan(&isExist); err != nil && !sqlx.IsErrNoRows(err) {
		return err
	}

	if isExist == 1 {
		return ecode.ErrDataAlreadyExists
	}

	now := time.Now().Unix()
	if err := p.db.Exec(insertClientSQL, 1, in.ID, now, in.Name, in.TokenSecret, in.TokenExpire, in.Credential); err != nil {
		return err
	}

	return nil
}

func (p *clientPersister) Update(_ context.Context, in *persistence.ClientUpdate) error {
	updateSQL := `UPDATE client SET updated_at = $1, %s WHERE id = $%d;`
	updateValuesTemplate := "%s = $%d"
	var updateValues []string

	var start = 1
	now := time.Now().Unix()
	args := []interface{}{now}
	if in.Name != nil {
		start++
		updateValues = append(updateValues, x.Sprintf(updateValuesTemplate, "name", start))
		args = append(args, *in.Name)
	}
	if in.TokenSecret != nil {
		start++
		updateValues = append(updateValues, x.Sprintf(updateValuesTemplate, "token_secret", start))
		args = append(args, *in.TokenSecret)
	}
	if in.TokenExpire != nil {
		start++
		updateValues = append(updateValues, x.Sprintf(updateValuesTemplate, "token_expire", start))
		args = append(args, *in.TokenExpire)
	}

	if start > 1 {
		start++
		args = append(args, in.ID)

		if err := p.db.Exec(x.Sprintf(updateSQL, strings.Join(updateValues, ", "), start), 1, args...); err != nil {
			return err
		}
	}

	return nil
}

func (p *clientPersister) Delete(_ context.Context, id string) error {
	return p.db.Exec("DELETE FROM client WHERE id = $1;", 1, id)
}
