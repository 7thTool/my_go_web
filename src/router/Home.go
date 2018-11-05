package router

import (
    "beego_server/src/controller"
    "github.com/astaxie/beego"
)

func init(){
    beego.Router("/",  &controller.HomeController{})
}
