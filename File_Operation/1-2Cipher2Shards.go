package main

import (
    //"bufio"
    "flag"
    "fmt"
    "os"
    "strconv"

    "Ruben"
)

/*
eg:
    go run 1-2Cipher2Shards.go \
                -file 1-1Cipher/8.txt \
                -size 2
 */


var cipherFile *string = flag.String("file", "Null", "Please input the cipherfile to be splited.")
var size       *string = flag.String("size", "0(kb)", "Please input the size of a shard.")

func main() {
    flag.Parse()
    if *cipherFile == "Null" {
        fmt.Println("no file to input")
        fmt.Print("Eg: go run test.go -file path/filename -size 2 (Kb) \n")
        return
    }else{
        file, err := os.Open(*cipherFile)
        if err != nil {
            fmt.Println("failed to open:", *cipherFile)
            fmt.Print("Eg: go run test.go -file path/filename -size 2 (Kb) \n")
        }else{
            defer file.Close()

            _, _, filenameOnly, fileSuffix := Ruben.DirFileNameSuffix(*cipherFile)
            size, _ := strconv.Atoi(*size)
            Folder := "1-2Shards/"
            Ruben.SplitFile(file, size*1024, Folder, filenameOnly, fileSuffix)
        }
    }
}
