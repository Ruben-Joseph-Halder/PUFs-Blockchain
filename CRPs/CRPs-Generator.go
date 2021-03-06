package main

import (
    "flag"
    "fmt"
    "os"
    "strconv"

    "strings"
    "Ruben"
)

/*
eg:
    go run CRPs-Generator.go \
                -file response_128_819200_0_220370.bin \
                -size 128

 */


var infile *string = flag.String("file", "Null", "Please input the challenge/response file.")
var size   *string = flag.String("size", "0(bit)", "Please input the size of the response.")
var PUFNum *string = flag.String("n", "Null", "Please input the PUFNum.")

func main() {
    flag.Parse() // Cipher/11.txt

    n,error := strconv.Atoi(*PUFNum)
    if error != nil{
        fmt.Println("Failed to convert string to integer")
    }

    if *infile == "Null" {
        fmt.Println("no file to input")
        fmt.Print("Eg: go run CRPs-Generator.go -file response.bin -size 128  (bit) \n")
        return
    }else{
        _, _, filenameOnly, fileSuffix := Ruben.DirFileNameSuffix(*infile)

        file, err := os.Open(*infile)
        if err != nil {
            fmt.Println("failed to open:", *infile)
            fmt.Print("Eg: go run test.go -file path/filename -size 2   (bit) \n")
        }else{
            defer file.Close()
            size, _ := strconv.Atoi(*size)

            if(strings.Index(filenameOnly, "challenge") > -1){
                Folder := "Challenge/"
                fmt.Println("get challenges...")
                Ruben.SplitFile(file, size*128, Folder, filenameOnly, fileSuffix)
            } else{
                Folder := "Response/"
                fmt.Println("get responses...")
                Ruben.SplitFile(file, size*8*n, Folder, filenameOnly, fileSuffix)

            }

        }

    }

}
