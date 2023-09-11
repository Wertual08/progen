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

func Generate(
    rootModule string,
    rootPath string,
    protoDirectory string,
    targetDirectory string,
) {
    log.Println("INFO: Cleaning up...")

    err := cleanup(rootPath, targetDirectory)
    if err != nil {
        log.Fatalf("FAIL: Can not list generated files\n%s", err)
    }
    
    err = generate(rootModule, rootPath, protoDirectory, targetDirectory)
    if err != nil {
        log.Fatalf("FAIL: Can not generate proto\n%s", err)
    }

    log.Println("INFO: Proto generated")
}

func cleanup(
    rootPath string,
    targetDirectory string,
) error {
    targetPath := filepath.Join(rootPath, targetDirectory)

    if _, err := os.Stat(targetPath); err != nil {
        if os.IsNotExist(err) {
            return nil
        }
        return err
    }

    err := filepath.WalkDir(
        targetPath,
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
    rootModule string,
    rootPath string, 
    protoDirectory string,
    targetDirectory string,
) error {
    protoPath := filepath.Join(rootPath, protoDirectory)
    targetPath := filepath.Join(rootPath, targetDirectory)

    argProtoPath := fmt.Sprintf(
        "--proto_path=%s", 
        protoPath,
    )
    argGoOut := fmt.Sprintf(
        "--go_out=paths=source_relative:%s", 
        targetPath,
    )
    argGoGrpcOut := fmt.Sprintf(
        "--go-grpc_out=paths=source_relative:%s", 
        targetPath,
    )

    targetModule := strings.Trim(filepath.ToSlash(targetDirectory), "/")

    argsOpt := make([]string, 0)

    err := filepath.WalkDir(
        protoPath,
        func(path string, d fs.DirEntry, err error) error {
            if err != nil { 
                return err
            }
            if !strings.HasSuffix(path, ".proto") {
                return nil
            }

            log.Printf("INFO: Collecting %s...\n", path)

            relativePath, err := filepath.Rel(protoPath, path)
            if err != nil {
                return err
            }
            
            fileModule := filepath.Dir(relativePath)
                
            argGoOpt := fmt.Sprintf(
                "--go_opt=M%s=%s/%s/%s", 
                relativePath, 
                rootModule,
                fileModule,
            )
            argGoGrpcOpt := fmt.Sprintf(
                "--go-grpc_opt=M%s=%s/%s/%s", 
                relativePath, 
                rootModule,
                targetModule,
                fileModule,
            )

            argsOpt = append(argsOpt, argGoOpt, argGoGrpcOpt)

            return nil
        },
    )
    if err != nil {
        return err
    }

    return filepath.WalkDir(
        protoPath,
        func(path string, d fs.DirEntry, err error) error {
            if err != nil { 
                return err
            }
            if !strings.HasSuffix(path, ".proto") {
                return nil
            }

            log.Printf("INFO: Generating %s...\n", path)

            args := append(
                argsOpt,
                argProtoPath,
                argGoOut,
                argGoGrpcOut,
                path,
            )

            cmd := exec.Command(
                "protoc", 
                args...,
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
}
