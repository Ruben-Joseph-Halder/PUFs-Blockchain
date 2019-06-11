/*
eg:
    go run File2Shards.go -file ./aa.pptx -size 1000
 */

package main

import (
    "fmt"
    "io/ioutil"
    "flag"

    "Ruben"
)

/*
eg:
    go run 2-2CipherFile.go \
                    -ciph 2-2MergeCipher/merge-8.txt \
                    -key File/8-cipherKey.txt

*/

var filePath *string = flag.String("ciph", "./MergeCipher/", "Please input the cipher file's path:")
var key      *string = flag.String("key", "Null", "Please input the secret key(16Byte): ")

func main(){
    flag.Parse()

    cipher, err := ioutil.ReadFile(*filePath)
    if err != nil {
        fmt.Print("err:", err)
    }else{
        fmt.Printf("\n%s %s","The file to be decrypted is:", *filePath)
        cipherText := []byte(cipher)

        keyplain, err := ioutil.ReadFile(*key)
        if err != nil {
            fmt.Print("err:", err)
        }else{
            key := []byte(keyplain)
            plain := Ruben.AesCBC_Decrypt(cipherText, key)

            if plain != nil {
                Folder := "File/"
                ExtraNmae := "decrypt-"
                Ruben.WriteInFile(Folder, ExtraNmae, *filePath, plain)
                fmt.Println("\n")
            }

        }

    }

}
