package main

import (
    "fmt"
    "io/ioutil"
    "flag"
    "Ruben"
)

/**
eg:
    go run 1-1File2Cipher-AES.go \
                -file File/8.txt \
                -key File/8-cipherKey.txt

**/

var filePath *string = flag.String("file", "Null", "Please input the file you want to encrpt: ")
var key      *string = flag.String("key", "Null", "Please input the secret key(16Byte): ")

func main(){
    flag.Parse()

    //2、读取文件内容
    plain, err := ioutil.ReadFile(*filePath)
    if err != nil {
        fmt.Print("err:", err)
        fmt.Print("\nEg: go run 1-1File2Cipher-AES.go -file File/11.txt \n")
    }else{
        fmt.Printf("\n %30s %s","The file to be encrpted is:", *filePath)
        plainText := []byte(plain)

        keyplain, err := ioutil.ReadFile(*key)
        if err != nil {
            fmt.Print("err:", err)
        }else{
            fmt.Printf("\n %30s %s","The key-file is:", *key)
            key := []byte(keyplain)

            cipherText := Ruben.AesCBC_Encrypt(plainText, key)

            if cipherText != nil {
                Folder := "1-1Cipher/"
                ExtraNmae := "cipher-"
                Ruben.WriteInFile(Folder, ExtraNmae, *filePath, cipherText)
            }

        }

    }

}
