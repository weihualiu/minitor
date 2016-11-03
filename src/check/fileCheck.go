
package check

import (
    "fmt"
    "time"
    "path/filepath"
    "os"
    "strings"
    "io"
    "crypto/sha1"
    "ftp"
    cf "config"
    "net/textproto"
    //"bytes"
    "bufio"
    "errors"
)

var fileHash map[string]string
var fileUpload map[string]bool

func init(){
	fileHash = make(map[string]string)
	fileUpload = make(map[string]bool)
}

func Checking() {
    
    for{
        time.Sleep(500 * time.Millisecond)
        //fmt.Println("output.....")
        recursionDirectory(cf.ConfigMap["check.directory"])
        processUpload()
    }

}

func listfunc(path string, f os.FileInfo, err error) error{
    var strRet string
    //strRet = path
    
    
    if f == nil{
        return err
    }
    if f.IsDir(){
        return nil
    }
    
    strRet += path
    
    //读取文件hash值
    file, errf := os.Open(strRet)
    defer file.Close()
    if errf != nil{
        fmt.Println("open file failed! filepath:",strRet)
        return nil
    }
    h := sha1.New()
    _, errf = io.Copy(h, file)
    if errf != nil{
        fmt.Println("sha1 generatre failed!:", strRet)
        return nil
    }
    strSha1 := string(h.Sum(nil)[:])
    
    //如果fileHash表中没有则添加至fileUpload中
    if val, ok := fileHash[strRet]; ok {
        //如果fileHash表中则比对是否一致，不一致则添加至fileUpload中
        if val != strSha1{
            fileUpload[strRet] = true
            fileHash[strRet] = strSha1
            //fmt.Println("sha1 str:",strSha1)
        }
    }else{
        
        fileUpload[strRet] = true
        fileHash[strRet] = strSha1
    }
    return nil
}

func recursionDirectory(dir string){
    filepath.Walk(dir, listfunc)
   
}

func processUpload(){
    files := make([]string, len(fileUpload), 100000)
    for key, _ := range fileUpload{
        // 处理upload文件
        fmt.Println("[", time.Now().Truncate(time.Second).String(), "] ", key)
        files = append(files, key)
    }
    if len(files) != 0 {
        upload(files)
    }
    
    fileUpload = make(map[string]bool)
}

// 上传不一致文件至ftp server
func upload(files []string) error{
    //连接ftp server
    hostStr := cf.ConfigMap["ftp.host"] + ":" + cf.ConfigMap["ftp.port"]
    c, err := ftp.DialTimeout(hostStr, 5*time.Second)
	if err != nil {
	    fmt.Println("connect failed! host addr:", hostStr, " err:", err)
		return err
	}
	/*if passive {
		delete(c.features, "EPSV")
	}*/
	err = c.Login(cf.ConfigMap["ftp.user"], cf.ConfigMap["ftp.password"])
	if err != nil {
	    fmt.Println("login failed!")
		return err
	}
    //切换目录
    err = c.ChangeDir(cf.ConfigMap["ftp.directory"])
	if err != nil {
	    fmt.Println("changedir:", cf.ConfigMap["ftp.directory"])
		return err
	}
	//fmt.Println("files len:", len(files))
    for _,filePath := range files {
        if filePath != "" {
            uploadFilePeer(c, filePath)
        }
        
    }
    //关闭连接
    err = c.Logout()
	if err != nil {
		if protoErr := err.(*textproto.Error); protoErr != nil {
			if protoErr.Code != ftp.StatusNotImplemented {
				return err
			}
		} else {
			return err
		}
	}

	c.Quit()
	
	return nil
}

/*
分析要上传的单个文件
path 相对路径
*/
func uploadFilePeer(conn *ftp.ServerConn,path string) error{
    //切换至配置指定的路径
    err := conn.ChangeDir(cf.ConfigMap["ftp.directory"])
    if err != nil {
        fmt.Println("changedir:", cf.ConfigMap["ftp.directory"])
        return err
    }
    
    //拆分前面无效串
    index := strings.Index(path, cf.ConfigMap["check.directory"])
    shortFileName := ""
    if index != -1 {
        shortFileName = strings.Replace(path, cf.ConfigMap["check.directory"], "", -1)
    }
    
    switchDir(conn, cf.ConfigMap["ftp.directory"], shortFileName)
    
    //上传文件
    file,err := os.Open(path)
    defer file.Close()
    if err != nil{
        return nil
    }
    data := bufio.NewReader(file)
    //data := bytes.NewBufferString(testData)
    
    // 取文件名称
    shortName := strings.Split(shortFileName, "\\")
    //fmt.Println("shortFileName:", shortName[len(shortName)-1])
	err = conn.Stor(shortName[len(shortName)-1], data)
	if err != nil {
		return err
	}
	
	return nil
}

/*
检查是否有此目录、创建目录、切换目录
*/
func switchDir(conn *ftp.ServerConn,before, after string) error {
    splitChar := "\\"
    
    if strings.Index(after, splitChar) == -1{
        //panic(errors.New("switch to dir failed!"))
        //fmt.Println("switchDir() not found ", splitChar)
        return nil
    }
    afterStr := strings.Split(after, splitChar)
    
    before1 := before
    if !strings.HasSuffix(before, splitChar) {
        before1 += "/"
    }
    before1 += afterStr[0]
    err := conn.ChangeDir(before1)
    if err != nil{
        fmt.Println("change to dir failed! ", before1)
        err := conn.MakeDir(before1)
        if err != nil{
            fmt.Println("mkdir dir failed! ", before1)
            panic(errors.New("switch to dir failed!"))
        }
    }
    err = conn.ChangeDir(before1)
    if err != nil{
        fmt.Println("change to dir failed! ", before1)
        panic(errors.New("switch to dir failed!"))
    }
    //如果长度是2则推出
    if len(afterStr) == 1{
        return nil
    }
    switchDir(conn, before1, strings.Join(afterStr[1:],"\\"))
    return nil
}