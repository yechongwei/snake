package dataconver

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	//ErrConverDataKindIsNotStruct 数据源不是struct类型
	ErrConverDataKindIsNotStruct = errors.New("=>data kind is not struct")
	//ErrConverDataKindIsNotMatch 源和目的数据kind类型不相同
	ErrConverDataKindIsNotMatch = errors.New("=>data kind is not match")
	//ErrConverDataKindIsInvalid 数据类型非法 既是源字段和目的字段名称没有匹配上
	ErrConverDataKindIsInvalid = errors.New("=>dst data kind is invalid")
	//ErrConverDataKindNotSupport 设置的数据源类型不支持
	ErrConverDataKindNotSupport = errors.New("=>data kind is not support")
)

//IConver struct结构转换接口
type IConver interface {
	//ConverData 转换数据接口  成功返回true 失败返回 false
	ConverData() (bool, error)
	//FindFieldByName 根据结构体字段名或则字段别名查找 反射的Value数据
	FindFieldByName(match func(filedName, tagName string) bool) reflect.Value
}

type Conver struct {
	srcStructName string
	srcTypeOf     reflect.Type
	srcValueOf    reflect.Value
	dstStructName string
	dstTypeOf     reflect.Type
	dstValueOf    reflect.Value
	tagLabel      string
}

//NewConver 创建数据转换类 传入的参数指针 Pointer:=&struct{}
//src  数据源结构体指针
//dest 数据目的结构体指针
//tagLabel 结构体字段取别名是的key:json `json:"legCount"`
func NewConver(src interface{}, dest interface{}, tagLabel string) IConver {
	c := &Conver{
		srcTypeOf:  reflect.TypeOf(src).Elem(),
		srcValueOf: reflect.ValueOf(src).Elem(),
		dstTypeOf:  reflect.TypeOf(dest).Elem(),
		dstValueOf: reflect.ValueOf(dest).Elem(),
		tagLabel:   tagLabel,
	}
	c.srcStructName = c.srcTypeOf.Name()
	c.dstStructName = c.dstTypeOf.Name()
	return c
}

//ConverData 转换数据接口
//成功返回true 失败返回 false
func (c *Conver) ConverData() (bool, error) {
	if c.srcTypeOf.Kind() != reflect.Struct {
		return false, ErrConverDataKindIsNotStruct
	}

	if c.dstTypeOf.Kind() != reflect.Struct {
		return false, ErrConverDataKindIsNotStruct
	}

	for i := 0; i < c.srcTypeOf.NumField(); i++ {
		//如果有别名的话 增加别名来处理
		srcFName, srcTName := c.srcTypeOf.Field(i).Name, c.srcTypeOf.Field(i).Tag.Get(c.tagLabel)
		//glog.Info("srcfieldName:", srcFName, " srctagName:", srcTName)
		//1.根据源字段名称或则别名 查找对应目的字段的Field Value()数据
		//如果源和目的字段顺序不一致的话 调用这个函数查找循环遍历FindFieldByName
		destFVal := c.FindFieldByNameByIndex(i, func(filedName, tagName string) bool {
			return srcFName == filedName || (srcTName != "" && srcTName == tagName)
		})

		//2.把源数据取出并且设置给目的数据对应的字段
		if ok, err := c.getAndSetValue(c.srcValueOf.Field(i), destFVal); !ok {
			return ok, err
		}
	}

	return true, nil
}

//FindFieldByNameByIndex 根据结构体字段名或则字段别名查找 反射的Value数据
//index:为结构体字段位置序号 这种方式需要2个结构体转换的字段名称|别名 一一按顺序对应
func (c *Conver) FindFieldByNameByIndex(index int, match func(filedName, tagName string) bool) reflect.Value {
	destFName, destTName := c.dstTypeOf.Field(index).Name, c.dstTypeOf.Field(index).Tag.Get(c.tagLabel)
	if match(destFName, destTName) {
		//glog.Info("match_name FName:", destFName, " TName:", destTName, " index:", index, " kind:", c.dstValueOf.Field(index).Kind())
		return c.dstValueOf.Field(index)
	}

	return reflect.Value{}
}

//FindFieldByName 根据结构体字段名或则字段别名查找 反射的Value数据
func (c *Conver) FindFieldByName(match func(filedName, tagName string) bool) reflect.Value {
	var val = reflect.Value{}
	for i := 0; i < c.dstTypeOf.NumField(); i++ {
		if val = c.FindFieldByNameByIndex(i, match); val.Kind() != reflect.Invalid {
			return val
		}
	}
	return val
}

func (c *Conver) getAndSetValue(srcFVal, destFVal reflect.Value) (ok bool, err error) {
	if destFVal.Kind() == reflect.Invalid {
		return false, fmt.Errorf("src:%s dst:%s %s", c.srcStructName, c.dstStructName, ErrConverDataKindIsInvalid)
	}

	if srcFVal.Kind() != destFVal.Kind() {
		return false, fmt.Errorf("src:%s dst:%s %s", c.srcStructName, c.dstStructName, ErrConverDataKindIsNotMatch)
	}

	ok = true
	switch k := srcFVal.Kind(); k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		destFVal.SetInt(srcFVal.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		destFVal.SetUint(srcFVal.Uint())
	case reflect.String:
		destFVal.SetString(srcFVal.String())
	case reflect.Float32, reflect.Float64:
		destFVal.SetFloat(srcFVal.Float())
	case reflect.Bool:
		destFVal.SetBool(srcFVal.Bool())
	case reflect.Slice:
		destFVal.SetBytes(srcFVal.Bytes())
	case reflect.Map:
		if srcFVal.Type() != destFVal.Type() {
			return false, fmt.Errorf("src:%s dst:%s %s: %s with %s", c.srcStructName, c.dstStructName, ErrConverDataKindIsNotMatch, srcFVal.Type(), destFVal.Type())
		}
		iter := srcFVal.MapRange()
		for iter.Next() {
			//glog.Info("key Kind:", iter.Key().Kind(), " value Kind:", iter.Value().Kind())
			destFVal.SetMapIndex(iter.Key(), iter.Value())
		}
	case reflect.Struct:
		con := NewConver(srcFVal.Addr().Interface(), destFVal.Addr().Interface(), c.tagLabel)
		if ok, err = con.ConverData(); !ok {
			return
		}
	default:
		ok, err = false, fmt.Errorf("%s Kind:%s", ErrConverDataKindNotSupport, k)

	}
	return
}
