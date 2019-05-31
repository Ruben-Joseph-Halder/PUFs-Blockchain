package main

import (
    "fmt"
    "os"
    "flag"
    "io/ioutil"
    "strconv"

    "Ruben"
)

/**
eg:
    go run 1-3ShardId_DataId.go \
                -ip 192.168.90.12 \
                -n 3 \
                -shard 1-2Shards/8.txt \
                -c CRPs/Challenge/challenge_128_819200_0_220370.bin \
                -r CRPs/Response/response_128_819200_0_220370.bin \
                -key File/cipherResponseKey.txt \
                -pub File/node_eccpublic.pem

**/
var IPaddress      *string = flag.String("ip",    "Null", "Please input the IPaddress: ")
var Number         *string = flag.String("n",     "0",    "Please input the ShardsNum: ")
var shardPath      *string = flag.String("shard", "Null", "Please input the file name: ")

var challengPath   *string = flag.String("c",     "Null", "Please input the challengPath: ")
var responsePath   *string = flag.String("r",     "Null", "Please input the responsePath: ")
var key            *string = flag.String("key",   "Null", "Please input the secret key1(16Byte): ")

var receiverPubKey *string = flag.String("pub",   "Null", "Please input the secret key1(16Byte): ")

func main(){
    flag.Parse()

    n,error := strconv.Atoi(*Number)
    if error != nil{
        fmt.Println("Failed to convert string to integer")
    }

    dir, _ := os.Getwd()
    shardDir,          _, shardNameOnly,              shardSuffix              := Ruben.DirFileNameSuffix(*shardPath)
    challengDir,       _, challengFilenameOnly,       challengFileSuffix       := Ruben.DirFileNameSuffix(*challengPath)
    responseDir,       _, responseFilenameOnly,       responseFileSuffix       := Ruben.DirFileNameSuffix(*responsePath)
    keyDir,            _, keyFilenameOnly,            keyFileSuffix            := Ruben.DirFileNameSuffix(*key)
    receiverPubKeyDir, _, receiverPubKeyFilenameOnly, receiverPubKeyFileSuffix := Ruben.DirFileNameSuffix(*receiverPubKey)


    for x := 0; x < n; x++ {
        shardName := shardNameOnly + "-" + strconv.Itoa(x) + shardSuffix
        filePath := dir + "\\" + shardDir + shardName
        ShardId := Ruben.ShardIdHash256(filePath, string(*IPaddress), x)
        hashFile := dir + "\\" + shardDir + "\\" + shardNameOnly + "-" + strconv.Itoa(x) + "-ShardId" + shardSuffix
        Ruben.WriteHashInFile(hashFile, ShardId)

        challeng, err := ioutil.ReadFile(dir + "\\" + challengDir + "\\" + challengFilenameOnly + "-" + strconv.Itoa(x) + challengFileSuffix)
        if err != nil {
            fmt.Print("err:", err)
            fmt.Print("\nEg: go run 1-4DataId.go -n 3 -c CRPs/Challenge/challenge_128_819200_0_220370-0.bin -key File/8-key.txt -r CRPs/Response/response_128_819200_0_220370-0.bin -sid 1-2Shards/8-0-ShardId.txt\n")
        } else{
            response, err := ioutil.ReadFile(dir + "\\" + responseDir + "\\" + responseFilenameOnly + "-" + strconv.Itoa(x) + responseFileSuffix)
            if err != nil {
                fmt.Print("err:", err)
            } else{
                plainText := []byte(response)
                keyplain, err := ioutil.ReadFile(dir + "\\" + keyDir + "\\" + keyFilenameOnly + "-" + strconv.Itoa(x) + keyFileSuffix)
                if err != nil {
                    fmt.Print("err:", err)
                }else{
                    fmt.Printf("\n%s %s","The key-file is:", dir + "\\" + keyDir + "\\"+ keyFilenameOnly + "-" + strconv.Itoa(x) + keyFileSuffix)
                    key := []byte(keyplain)

                    cipherResponse := Ruben.AesCBC_Encrypt(plainText, key)
                    if cipherResponse != nil{
                        data := string(challeng) + string(cipherResponse) + ShardId
                        DataId := Ruben.DataIdHash256(data)

                        DataIdfile := dir + "\\" + shardDir + "\\" + shardNameOnly + "-" + strconv.Itoa(x) + "-DataId" + shardSuffix
                        Ruben.WriteHashInFile(DataIdfile, DataId)


                        plainCRPs := string(challeng) + string(cipherResponse)
                        eccpublic, err := ioutil.ReadFile(dir + "\\" + receiverPubKeyDir + "\\" + receiverPubKeyFilenameOnly + "-" + strconv.Itoa(x) + receiverPubKeyFileSuffix)
                        fmt.Printf("\n%s %s","The key-file is:", dir + "\\" + receiverPubKeyDir + "\\" + receiverPubKeyFilenameOnly + "-" + strconv.Itoa(x) + receiverPubKeyFileSuffix)
                        if err != nil {
                            fmt.Print("err:", err)
                        } else{
                            publicKey:=[]byte(string(eccpublic))
                            cipherCRPs,_:=Ruben.EccPublicEncrypt([]byte(plainCRPs),publicKey)

                            cipherCRPsPath := dir + "\\" + shardDir + "\\" + receiverPubKeyFilenameOnly + "-" + strconv.Itoa(x) +  "-cipherCRPs" + challengFileSuffix
                            Ruben.WriteHashInFile(cipherCRPsPath, string(cipherCRPs))
                            fmt.Println("\n")

                        }

                    }

                }

            }

        }

    }

}



