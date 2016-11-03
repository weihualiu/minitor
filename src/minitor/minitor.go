package main

import(
    "fmt"
    "os"
    cf "config"
    ck "check"
)


func main(){
    fmt.Println("*************created by liu.weihua  ver:1.0**************")
    fmt.Println("minitor starting")
    configPath := "minitor.conf"
    if len(os.Args) == 2{
        configPath = os.Args[1]
    }
    cf.InitConfig(configPath)
    for key,_ := range cf.ConfigMap{
        fmt.Println("config :",key, "=" ,cf.ConfigMap[key])
    }
    
    ck.Checking()
}

