package main

import (
	_ "beego_server/src/router"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"syscall"

	"html/template"

	"net/http"
	"strings"
	//"github.com/astaxie/beego"
)

func execShell(s string) {
	cmd := exec.Command("/bin/bash", "-c", s)
	var out bytes.Buffer

	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s", out.String())
}

func exeCmd(cmd string, wg *sync.WaitGroup) {
	fmt.Println(cmd)
	parts := strings.Fields(cmd)
	out, err := exec.Command(parts[0], parts[1]).Output()
	if err != nil {
		fmt.Println("error occured")
		fmt.Printf("%s", err)
	}
	fmt.Printf("%s", out)
	wg.Done()
}

type Service struct {
	// Other things
	ch        chan bool
	waitGroup *sync.WaitGroup
}

func NewService() *Service {
	s := &Service{
		// Init Other things
		ch:        make(chan bool),
		waitGroup: &sync.WaitGroup{},
	}
	return s
}

func (s *Service) Stop() {
	close(s.ch)
	s.waitGroup.Wait()
}

func (s *Service) Serve1() {
	s.waitGroup.Add(1)
	defer s.waitGroup.Done()

	for {
		select {
		case <-s.ch:
			fmt.Println("stopping...")
			return
		default:
		}
		go s.Serve2()
	}
}

func (s *Service) Serve2() {
	s.waitGroup.Add(1)
	defer s.waitGroup.Done()
	for {
		select {
		case <-s.ch:
			fmt.Println("stopping...")
			return
		default:
		}
		// Do something
	}
}

// 崩溃时需要传递的上下文信息
type panicContext struct {
	function string // 所在函数
}

// 保护方式允许一个函数
func ProtectRun(entry func()) {
	// 延迟处理的函数
	defer func() {
		// 发生宕机时，获取panic传递的上下文并打印
		err := recover()
		switch err.(type) {
		case runtime.Error: // 运行时错误
			fmt.Println("runtime error:", err)
		default: // 非运行时错误
			fmt.Println("error:", err)
		}
	}()
	entry()
}

func sayHello(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()       //解析参数，默认是不会解析的
	fmt.Println(r.Form) //这些信息是输出到服务器端的打印信息
	fmt.Println("path", r.URL.Path)
	fmt.Println("scheme", r.URL.Scheme)
	fmt.Println(r.Form["url_long"])
	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, ""))
	}
	fmt.Fprintf(w, "Hello astaxie!") //这个写入到w的是输出到客户端的
}

func login(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method) //获取请求的方法
	if r.Method == "GET" {
		t, _ := template.ParseFiles("login.gtpl")
		t.Execute(w, nil)
	} else {
		r.ParseForm() //解析参数，默认是不会解析的
		//请求的是登陆数据，那么执行登陆的逻辑判断
		fmt.Println("username:", r.Form["username"])
		fmt.Println("password:", r.Form["password"])
	}
}

func closeServer(w http.ResponseWriter, r *http.Request) {
	//panic("closeServer")
	execShell("kill -9 " + strconv.Itoa(os.Getpid()))
}

func run() {
	//设置访问的路由
	http.HandleFunc("/", sayHello)
	http.HandleFunc("/login", login)
	http.HandleFunc("/closeServer", closeServer)
	err := http.ListenAndServe(":9092", nil) //设置监听的端口
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

	//beego.Run()
}

func main() {
	execShell("uname ")

	wg := new(sync.WaitGroup)
	commands := []string{"echo newline >> foo.o", "echo newline >> f1.o", "echo newline >> f2.o"}
	for _, str := range commands {
		wg.Add(1)
		go exeCmd(str, wg)
	}
	wg.Wait()

	exit := make(chan os.Signal)                         //初始化一个channel
	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM) //notify方法用来监听收到的信号

	//service := NewService()
	//go service.Serve1()
	//go service.Serve2()
	go run()

	sig := <-exit
	fmt.Println("exit:", sig.String())

	// Stop the service gracefully.
	//service.Stop()

	fmt.Println("运行前")
	// 允许一段手动触发的错误
	ProtectRun(func() {
		fmt.Println("手动宕机前")
		// 使用panic传递上下文
		panic(&panicContext{
			"手动触发panic",
		})
		fmt.Println("手动宕机后")
	})
	// 故意造成空指针访问错误
	ProtectRun(func() {
		fmt.Println("赋值宕机前")
		var a *int
		*a = 1
		fmt.Println("赋值宕机后")
	})
	fmt.Println("运行后")
}
