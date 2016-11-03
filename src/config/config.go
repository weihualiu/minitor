package config

import(
    "fmt"
    "os"
    "bufio"
    "strings"
    "io"
)

var ConfigMap map[string]string


func init(){
    ConfigMap = make(map[string]string, 100000)
}

func InitConfig(configPath string) error{
    file, err := os.Open(configPath)
    defer file.Close()
    
    if err != nil{
        fmt.Println("read config file error!")
        return err
    }
    buf := bufio.NewReader(file)
    for{
        line, err := buf.ReadString('\n')
        configParser(line)
        if err != nil{
            if err == io.EOF{
                return nil
            }
            return err
        }
    }
}

func configParser(row string) error{
    str := strings.TrimSpace(row)
    //判断是否是#开头
    if strings.HasPrefix(str, "#"){
        //fmt.Println("not context:", str)
        return nil
    }
    // 去掉#之后的内容
    str1 := str
    if strings.Contains(str, "#") {
        str1 = strings.Split(str, "#")[0]
    }
    //根据=拆分
    if !strings.Contains(str1, "=") {
        return nil
    }
    str2 := strings.Split(str1, "=")
    //去掉两端空格
    ConfigMap[strings.TrimSpace(str2[0])] = strings.TrimSpace(str2[1])
    
    return nil
}