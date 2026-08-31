package common

import (
	"database/sql/driver"
	"fmt"
	"uuid"
)

type UUID [16]byte

func NewUUIDv7() UUID {
	return UUID(uuid.NewV7())
}

func MustUUIDFromString(s string) UUID {
	return UUID(uuid.MustParse(s))
}

func (u UUID) String() string {
	return uuid.UUID(u).String()
}

func (u UUID) MarshalText() ([]byte, error) {
	return uuid.UUID(u).MarshalText()
}

func (u *UUID) UnmarshalText(data []byte) error {
	var parsed uuid.UUID
	if err := parsed.UnmarshalText(data); err != nil {
		return err
	}

	*u = UUID(parsed)
	return nil
}

func (u UUID) Value() (driver.Value, error) {
	return u.String(), nil
}

func (u *UUID) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		return nil
	case string:
		if v == "" {
			return nil
		}

		parsed, err := uuid.Parse(v)
		if err != nil {
			return fmt.Errorf("Scan: %w", err)
		}

		*u = UUID(parsed)
		return nil
	case []byte:
		if len(v) == 0 {
			return nil
		}

		if len(v) != 16 {
			return u.Scan(string(v))
		}

		copy(u[:], v)
		return nil
	default:
		return fmt.Errorf("Scan: unable to scan type %T into UUID", src)
	}
}

func (u UUID) IsZero() bool {
	return u == UUID(uuid.Nil())
}

func (u UUID) Equals(other UUID) bool {
	return u == other
}
