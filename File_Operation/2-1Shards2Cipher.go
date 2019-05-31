package main

import (
    "fmt"
    "os"
    "flag"

    "Ruben"
)

/*
eg:
    go run 2-1Shards2Cipher.go \
                    -fold 2-1Retrieve/8.txt

*/

var filePath *string = flag.String("fold", "2-1Retrieve", "Please input the folder that contain the shards:")

func main() {
    flag.Parse()

    dir, _ := os.Getwd()
    fileDir, _, filenameOnly, fileSuffix := Ruben.DirFileNameSuffix(*filePath)
    rootPath := dir + "\\" + fileDir
    fmt.Println("\nThe folder that contains the shards is: ", rootPath)
    Ruben.MergeFile(rootPath, filenameOnly, fileSuffix)

}

