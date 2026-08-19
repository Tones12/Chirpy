package auth

import (
	"fmt"
	"testing"
	"time"
    "github.com/google/uuid"
    "net/http"
)

func TestHashPassword(t *testing.T) {
    name := "password"
    fmt.Printf("\nTesting pw: %q\n", name)
    hash, err := HashPassword(name)
    if err != nil {
       t.Errorf(`Error hashing %q, %v`, name, err)
    }
    fmt.Printf("Produced hash = %q\n\n", hash)

    fmt.Println("Testing check password hash function:")

    ok, err := CheckPasswordHash(name, hash)
    if ok == false || err != nil {
        fmt.Printf("Password check failed, %v\n\n", err)
    }
    fmt.Println("Password check succeded")
}

func TestJWTNormal(t *testing.T) {
    userID := uuid.New()
    tokenSecret := "password"
    expiresIn, err := time.ParseDuration("5m")
    fmt.Println("\nTesting token creation:")
    
    jwtString, err := MakeJWT(userID, tokenSecret, expiresIn)
    if err != nil {
       t.Errorf(`Error making JWT %q, %v`, jwtString, err)
    }
    fmt.Printf("JWT token string = %q\n\n", jwtString)

    fmt.Println("Testing token validation:")

    _, err = ValidateJWT(jwtString, tokenSecret)
    if err != nil {
        fmt.Printf("Token validation failed, %v\n\n", err)
    }
    fmt.Println("Token validation succeded")
}

func TestJWTExpired(t *testing.T) {
    userID := uuid.New()
    tokenSecret := "password"
    expiresIn, err := time.ParseDuration("-5s")
    fmt.Println("\nTesting expired token:")
    
    jwtString, err := MakeJWT(userID, tokenSecret, expiresIn)
    if err != nil {
       t.Errorf(`Error making JWT %q, %v`, jwtString, err)
    }

    _, err = ValidateJWT(jwtString, tokenSecret)
    if err != nil {
        fmt.Printf("Token validation failed, %v\n\n", err)
    }
    if err == nil {
        t.Fatal("expected an error")
    }
}

func TestJWTWrongSecret(t *testing.T) {
    userID := uuid.New()
    tokenSecret := "password"
    expiresIn, err := time.ParseDuration("-5s")
    fmt.Println("\nTesting an incorrect secret password:")
    
    jwtString, err := MakeJWT(userID, tokenSecret, expiresIn)
    if err != nil {
       t.Errorf(`Error making JWT %q, %v`, jwtString, err)
    }

    _, err = ValidateJWT(jwtString, "wrongpassword")
    if err != nil {
        fmt.Printf("Token validation failed, %v\n\n", err)
    }
    if err == nil {
        t.Fatal("expected an error")
    }
}

func TestJWTWrongTokenString(t *testing.T) {
    userID := uuid.New()
    tokenSecret := "password"
    expiresIn, err := time.ParseDuration("-5s")
    fmt.Println("\nTesting wrong token string:")
    
    jwtString, err := MakeJWT(userID, tokenSecret, expiresIn)
    if err != nil {
       t.Errorf(`Error making JWT %q, %v`, jwtString, err)
    }

    _, err = ValidateJWT("wrong string", tokenSecret)
    if err != nil {
        fmt.Printf("Token validation failed, %v\n\n", err)
    }
    if err == nil {
        t.Fatal("expected an error")
    }
}

func TestGetBearerToken(t *testing.T) {
    headers := http.Header{}
    headers.Set("Authorization", "Bearer randomTokenString1")
    token, err := GetBearerToken(headers)
    if err != nil {
        fmt.Printf("expected no error, got %v", err)
    }
    if token != "randomTokenString" {
        fmt.Printf("expected 'randomTokenString', got '%s'", token)
    }
}