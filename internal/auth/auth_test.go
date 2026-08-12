package auth

import (
    "testing"
    "fmt"
)

// TestHelloName calls greetings.Hello with a name, checking
// for a valid return value.
func TestHashPassword(t *testing.T) {
    name := "pa$$w0rd"
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