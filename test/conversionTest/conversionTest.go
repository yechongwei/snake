package main

import (
	"dataconver"
	"flag"
	"fmt"
	"reflect"
	"time"

	"github.com/golang/glog"
)

type people struct {
	name string
	age  uint32
}

type StructField struct {
	Name      string  // 字段名
	PkgPath   string  // 字段路径
	Type      int     // 字段反射类型对象
	Tag       string  // 字段的结构体标签
	Offset    uintptr // 字段在结构体中的相对偏移
	Index     []int   // Type.FieldByIndex中的返回的索引值
	Anonymous bool    `json:"type" id:"100"` // 是否为匿名字段
}

func testTypeOf() {
	var person = &people{}
	typeOfA := reflect.TypeOf(person.age)
	fmt.Println(typeOfA.Name(), ",", typeOfA.Kind())

	typeOfA = reflect.TypeOf(person)
	fmt.Println(typeOfA.Name(), ",", typeOfA.Kind())
	typeOfA = typeOfA.Elem()
	fmt.Println(typeOfA.Name(), ",", typeOfA.Kind())

	ins := StructField{Name: "yezongwei"}
	// 获取结构体实例的反射类型对象
	typeOfField := reflect.TypeOf(ins)

	fmt.Println(typeOfField.Name(), ",", typeOfField.Kind())

	for i := 0; i < typeOfField.NumField(); i++ {
		// 获取每个成员的结构体字段类型
		fieldType := typeOfField.Field(i)
		// 输出成员名和tag
		fmt.Printf("name: %v  tag: '%v'  PkgPath:'%v' \n", fieldType.Name, fieldType.Tag, fieldType.PkgPath)
	}
}

// 定义结构体
type dummy struct {
	a int
	b string
	// 嵌入字段
	float32
	bool
	next *dummy
}

func testValueOf() {
	//1
	fmt.Println("==================test ValueOf  1==================")
	var a int = 1024
	// 获取变量a的反射值对象
	valueOfA := reflect.ValueOf(a)
	// 获取interface{}类型的值, 通过类型断言转换
	var getA int = valueOfA.Interface().(int)
	// 获取64位的值, 强制类型转换为int类型
	var getA2 int = int(valueOfA.Int())
	fmt.Println(getA, getA2)

	//2
	fmt.Println("==================test ValueOf  2  struct===============")
	// 值包装结构体
	d := reflect.ValueOf(dummy{
		a:    1,
		b:    "yezongwei",
		next: &dummy{},
	})
	// 获取字段数量
	fmt.Println("NumField", d.NumField())
	// 获取索引为2的字段(int字段)
	Field := d.Field(1)
	// 输出字段类型
	fmt.Println("Field", Field.Type(), "FieldName:", Field.String())
	// 根据名字查找字段
	fmt.Println("FieldByName(\"b\").Type", d.FieldByName("b").Type())
	// 根据索引查找值中, next字段的int字段的值
	fmt.Println("FieldByIndex([]int{4, 0}).Type()", d.FieldByIndex([]int{4, 4}).Type())
}

type inner struct {
	Inner int
}

type color struct {
	BodyColor int8   //颜色
	LegColor  string //颜色
	Extends   map[string]map[string]string
	Inner     inner
}
type dog struct {
	LegCount int `json:"legCount"`
	Name     string
	Score    float64
	Age      int16 `json:"age"`
	B        bool
	Info     []byte
	Infos    map[int]map[int]string
	Color    color
}

type dog2 struct {
	DogLegCount int `json:"legCount"`
	Name        string
	Score       float64
	DogAge      int16 `json:"age"`
	B           bool
	Info        []byte
	Infos       map[int]map[int]string
	Color       color
}

func testSetValue2() {
	ins := &dog{
		LegCount: 4,
		Name:     "dog wawa",
		Score:    1.23456,
		Age:      10,
		B:        true,
		Info:     []byte{'k', 'i', 't'},
		Infos: map[int]map[int]string{
			0: {
				0: "hello",
				1: "world",
			},
			1: {
				0: "hello1",
				1: "world1",
			},
		},
		Color: color{
			BodyColor: 1,
			LegColor:  "red",
			Extends: map[string]map[string]string{
				"why": {
					"say":  "hello",
					"call": "world",
				},
				"not": {
					"eat":    "meat",
					"carray": "on",
				},
			},
			Inner: inner{100},
		},
	}

	ins1 := &dog2{
		Infos: make(map[int]map[int]string),
		Color: color{
			Extends: make(map[string]map[string]string),
			Inner:   inner{300},
		},
	}

	now := time.Now()
	c := dataconver.NewConver(ins, ins1, "json")
	//var err error
	_, err := c.ConverData()
	offTime := time.Now().Sub(now)
	glog.Info(err, " offTime:", offTime)
	glog.Info(ins)
	glog.Info(ins1)

	// ins = &dog{
	// 	LegCount: 4,
	// 	Name:     "dog wawa *******",
	// 	Score:    1.23456,
	// 	Age:      10,
	// 	B:        true,
	// 	Info:     []byte{'d', 'o', 'g'},
	// 	Infos: map[int]map[int]string{
	// 		0: {
	// 			0: "hello",
	// 			1: "kit",
	// 		},
	// 		1: {
	// 			0: "hello1",
	// 			1: "kit1",
	// 		},
	// 	},
	// }
	// now = time.Now()
	// ins1.DogLegCount = ins.LegCount
	// ins1.DogAge = ins.Age
	// ins1.B = ins.B
	// ins1.Info = ins.Info
	// ins1.Score = ins.Score
	// ins1.Name = ins.Name
	// ins1.Infos = ins.Infos
	// offTime = time.Now().Sub(now)
	// glog.Info(err, " offTime:", offTime)
	// glog.Info(ins)
	// glog.Info(ins1)
}

func testReflectNil() {
	// *int的空指针
	var a *int
	fmt.Println("var a *int:", reflect.ValueOf(a).IsNil())
	// nil值
	fmt.Println("nil:", reflect.ValueOf(nil).IsValid())
	// *int类型的空指针
	fmt.Println("(*int)(nil):", reflect.ValueOf((*int)(nil)).Elem().IsValid())
	// 实例化一个结构体
	s := struct{}{}
	// 尝试从结构体中查找一个不存在的字段
	fmt.Println("不存在的结构体成员:", reflect.ValueOf(s).FieldByName("").IsValid())
	// 尝试从结构体中查找一个不存在的方法
	fmt.Println("不存在的结构体方法:", reflect.ValueOf(s).MethodByName("").IsValid())
	// 实例化一个map
	m := map[int]int{}
	// 尝试从map中查找一个不存在的键
	fmt.Println("不存在的键：", reflect.ValueOf(m).MapIndex(reflect.ValueOf(3)).IsValid())
}

func init() {
	glog.MaxSize = 1024 * 1024 * 100    //最大100M
	flag.Set("alsologtostderr", "true") // 日志写入文件的同时，输出到stderr
	flag.Set("log_dir", "./log")        // 日志文件保存目录
	flag.Set("v", "1")                  // 配置V输出的等级。
	flag.Parse()
}

func main() {
	//testTypeOf()
	//testValueOf()
	//testReflectNil()
	testSetValue2()
}
