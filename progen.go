package progen

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Run(
    rootPath string,
    protoDirectory string,
    outputDirectory string,
) {
    log.Println("INFO: Cleaning up...")

    err := cleanup(rootPath, outputDirectory)
    if err != nil {
        log.Fatalf("FAIL: Can not list generated files\n%s", err)
        os.Exit(1)
    }
    
    err = generate(rootPath, protoDirectory, outputDirectory)
    if err != nil {
        log.Fatalf("FAIL: Can not generate proto\n%s", err)
        os.Exit(2)
    }

    log.Println("INFO: Proto generated")
}

func cleanup(
    rootPath string,
    outputDirectory string,
) error {
    err := filepath.WalkDir(
        filepath.Join(rootPath, outputDirectory),
        func(path string, d fs.DirEntry, err error) error {
            if err != nil {
                return err
            }
            if !strings.HasSuffix(path, ".pb.go") {
                return nil
            }

            return os.Remove(path)
        },
    )
    if err != nil {
        return err
    }


    return nil
}

func generate(
    rootPath string, 
    protoDirectory string,
    outputDirectory string,
) error {
    targetProtoDirectory := filepath.Join(rootPath, protoDirectory)

    argProtoPath := fmt.Sprintf(
        "--proto_path=%s/%s", 
        rootPath, 
        protoDirectory,
    )
    argGoOut := fmt.Sprintf(
        "--go_out=%s", 
        rootPath,
    )
    argGoGrpcOut := fmt.Sprintf(
        "--go-grpc_out=%s", 
        rootPath,
    )

    err := filepath.WalkDir(
        targetProtoDirectory,
        func(path string, d fs.DirEntry, err error) error {
            if err != nil {
                return err
            }
            if !strings.HasSuffix(path, ".proto") {
                return nil
            }

            log.Printf("INFO: Generating %s...\n", path)

            relativePath := strings.TrimPrefix(path, targetProtoDirectory)[1:]

            fileModule := strings.TrimSuffix(
                relativePath,
                ".proto",
            )
                
            argGoOpt := fmt.Sprintf(
                "--go_opt=M%s=%s/%s", 
                relativePath, 
                outputDirectory, 
                fileModule,
            )
            argGoGrpcOpt := fmt.Sprintf(
                "--go-grpc_opt=M%s=%s/%s", 
                relativePath, 
                outputDirectory, 
                fileModule,
            )

            cmd := exec.Command(
                "protoc", 
                argProtoPath,
                argGoOut,
                argGoGrpcOut,
                argGoOpt,
                argGoGrpcOpt,
                path,
            )

            if _, err := cmd.Output(); err != nil {
                if err, ok := err.(*exec.ExitError); ok {
                    return errors.Join(err, errors.New(string(err.Stderr)))
                }

                return err
            }

            return nil
        },
    )
    if err != nil {
        return err
    }

    return nil
}
