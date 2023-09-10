package progen

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func Run(
    rootDirectory string,
    outputDirectory string,
) {
    log.Println("INFO: Cleaning up...")

    err := cleanup(rootDirectory)
    if err != nil {
        log.Fatalf("FAIL: Can not list generated files\n%s", err)
        os.Exit(1)
    }
    
    err = generate(rootDirectory, outputDirectory)
    if err != nil {
        log.Fatalf("FAIL: Can not generate proto\n%s", err)
        os.Exit(2)
    }

    log.Println("INFO: Proto generated")
}

func cleanup(rootDirectory string) error {
    matches, err := filepath.Glob(
        fmt.Sprintf("%s/*/*.pb.go", rootDirectory),
    )
    if err != nil {
        return err
    }

    for _, match := range matches {
        if err := os.Remove(match); err != nil {
            return err
        }
    }

    return nil
}

func generate(rootDirectory string, outputDirectory string) error {
    matches, err := filepath.Glob(
        fmt.Sprintf("%s/*/*.proto", rootDirectory),
    )
    if err != nil {
        return err
    }

    protoPath := fmt.Sprintf("--proto_path=%s/proto", rootDirectory)
    goOut := fmt.Sprintf("--go_out=%s", outputDirectory)
    goGrpcOut := fmt.Sprintf("--go-grpc_out=%s", outputDirectory)
    
    for _, match := range matches {
        log.Printf("INFO: Generating %s...\n", match)

        cmd := exec.Command(
            "protoc", 
            protoPath,
            match,
            goOut,
            goGrpcOut,
        )

        if _, err := cmd.Output(); err != nil {
            if err, ok := err.(*exec.ExitError); ok {
                return errors.Join(err, errors.New(string(err.Stderr)))
            }

            return err
        }
    }

    return nil
}
