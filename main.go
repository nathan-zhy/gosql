package main

import (
	"database/sql"
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

func GetOutBoundIP() (ip string, err error) {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		fmt.Println(err)
		return
	}
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	fmt.Println(localAddr.String())
	ip = strings.Split(localAddr.String(), ":")[0]
	return
}

type SysWexinUserCode struct {
	ID           int64
	Openid       string
	Code         string
	Order_no     string
	Created_by   string
	Phone        sql.NullString
	Created_time sql.NullString
	Update_by    sql.NullString
	Update_time  sql.NullString
}

func main() {

	sysType := runtime.GOOS

	if sysType == "linux" {
		gin.SetMode(gin.ReleaseMode)
	}
	ip, err := GetOutBoundIP()
	if err != nil {
		fmt.Println(err)
	}
	if ip == "172.24.191.33" {
		err := godotenv.Load("dev.env")
		if err != nil {
			fmt.Println(err)
		}
	} else {
		err := godotenv.Load("dev.envtest")
		if err != nil {
			fmt.Println(err)
		}
	}

	dsn := os.Getenv("DATABASE_URL")
	fmt.Printf("db on:%s", dsn)
	//使用connect连接数据库
	db := sqlx.MustOpen("mysql", dsn)
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(20)

	r := gin.Default()
	r.POST("/account/purchase", func(context *gin.Context) {
		json := make(map[string]map[string]string)
		context.BindJSON(&json)
		var user SysWexinUserCode
		select_query := fmt.Sprintf("SELECT * FROM sys_wexin_user_code WHERE phone='%s'", json["set"]["phone"])
		fmt.Println(select_query)
		if err := db.Get(&user, select_query); err != nil {
			context.JSON(200, gin.H{"use": fmt.Sprintf("\"%s already have a code\"", json["set"]["phone"])})
			return
		}
		select_query = fmt.Sprintf("SELECT * FROM sys_wexin_user_code WHERE code='%s'", json["filter"]["code"])
		fmt.Println(select_query)
		if err := db.Get(&user, select_query); err != nil {
			fmt.Printf("get failed, err:%v\n", err)
			return
		}
		if user.Phone.Valid != false {
			context.JSON(200, gin.H{"use": fmt.Sprintf("\"already used by %s\"", user.Phone.String)})
			return
		}
		select_query = fmt.Sprintf("UPDATE sys_wexin_user_code SET phone='%s' WHERE id='%d'", json["set"]["phone"], user.ID)
		fmt.Println(select_query)
		if _, err := db.Exec(select_query); err != nil {
			fmt.Printf("get failed, err:%v\n", err)
			return
		}

		context.JSON(200, gin.H{"use": "1"})
	})

	fmt.Println("start server on:17733")
	r.Run(":17733")
}
