package auth

import "github.com/ranggadablues/gosok/auth/jwt"

var manager *jwt.Manager

func Init() error {
	m, err := jwt.NewManagerFromEnv()
	if err != nil {
		return err
	}
	manager = m
	return nil
}

func Manager() *jwt.Manager {
	return manager
}
