package golist

import (
	"fmt"
)

//IList 接口类
type IList interface {
	Len() uint32
	IsEmpty() bool
	Push(index uint32, data interface{}) bool
	RPush(data interface{}) bool
	LPush(data interface{}) bool
	Pop(index uint32) interface{}
	RPop() interface{}
	LPop() interface{}
	//匹配value值相同的 interface{}:为节点的存的值 返回匹配的节点位置和是否查找到 如果匹配不到则返回的第一个参数index无效
	Match(key interface{}, fn func(key, value interface{}) bool) (uint32, bool)
	//匹配value值相同的interface{}:为节点的存的值 并且返回是否找到并且删除
	MatchAndRemove(key interface{}, fn func(key, value interface{}) bool) bool
}

//ListNode 链表节点
//value 节点存放的数据
//prev 上一个节点
//next  下一个节点
type ListNode struct {
	value interface{}
	prev  *ListNode
	next  *ListNode
}

//GoList 链表  多线程不安全 外层需要加锁保护
//head 链表头节点
//tail 链表尾节点
//len 链表长度
type GoList struct {
	head   *ListNode
	tail   *ListNode
	len    uint32
	getLen uint64
	putLen uint64
}

func newNode(data interface{}, prev, next *ListNode) *ListNode {
	return &ListNode{
		value: data,
		prev:  prev,
		next:  next,
	}
}

//NewList 创建一个链表
func NewList() IList {
	return &GoList{
		head: nil,
		tail: nil,
		len:  0,
	}
}

func (l *GoList) String() string {
	return fmt.Sprintln("len:", l.len, " putLen:", l.putLen, " getLen:", l.getLen)
}

//Len  获取链表长度
func (l *GoList) Len() uint32 {
	return l.len
}

//IsEmpty 链表是否为空
func (l *GoList) IsEmpty() bool {
	return l.len == 0
}

//Push 往链表固定位置存放数据
//存放成功返回true  失败返回false
func (l *GoList) Push(index uint32, data interface{}) bool {
	if index == 0 {
		return l.LPush(data)
	}

	if index >= l.len {
		return l.RPush(data)
	}

	//在区间左半边 则从头往尾遍历插入 [ 1 - len/2]
	if index <= l.len>>1 {
		return l.lPushByIndex(index, data)
	}
	//在区间右半边 则从尾往头遍历插入 (len/2 - len -1]
	return l.rPushByIndex(index, data)
}

//lPushByIndex 如果插入位置在左半边 则从头往尾遍历寻找插入
func (l *GoList) lPushByIndex(index uint32, data interface{}) bool {
	nodeTmp := l.head
	for i := uint32(1); i <= l.len>>1; i++ {
		if index == i {
			if node := newNode(data, nodeTmp, nodeTmp.next); node != nil {
				nodeTmp.next.prev = node
				nodeTmp.next = node
				l.len++
				l.putLen++
				return true
			} else {
				break
			}
		}
		nodeTmp = nodeTmp.next
	}

	return false
}

//rPushByIndex 如果插入位置在右半边 则从尾往头遍历寻找插入
func (l *GoList) rPushByIndex(index uint32, data interface{}) bool {
	nodeTmp := l.tail
	for i := l.len - 1; i > l.len>>1; i-- {
		if index == i {
			if node := newNode(data, nodeTmp.prev, nodeTmp); node != nil {
				nodeTmp.prev.next = node
				nodeTmp.prev = node
				l.len++
				l.putLen++
				return true
			} else {
				break
			}
		}
		nodeTmp = nodeTmp.prev
	}

	return false
}

//RPush  往链表尾部后插入
func (l *GoList) RPush(data interface{}) bool {
	if node := newNode(data, l.tail, nil); node != nil {
		if l.tail == nil {
			l.head = node
		} else {
			l.tail.next = node
		}
		l.tail = node
		l.len++
		l.putLen++
		return true
	}
	return false
}

//LPush 往链表头部前插入
func (l *GoList) LPush(data interface{}) bool {
	if node := newNode(data, nil, l.head); node != nil {
		if l.head == nil {
			l.tail = node
		} else {
			l.head.prev = node
		}
		l.head = node
		l.len++
		l.putLen++
		return true
	}
	return false
}

//Pop 从链表固定位置取数据
//成功返回数据  失败返回nil
func (l *GoList) Pop(index uint32) interface{} {
	if l.len == 0 {
		return nil
	}

	if index == 0 {
		return l.LPop()
	}

	if index >= l.len-1 {
		return l.RPop()
	}

	//在区间左半边 则从头往尾遍历取 [1 - len/2)
	if index < l.len>>1 {
		return l.lPopByIndex(index)
	}
	//在区间右半边 则从尾往头遍历取 [len/2 - len-2]
	return l.rPopByIndex(index)
}

//lPopByIndex 如果位置在左半边 则从头往尾遍历寻找取出
func (l *GoList) lPopByIndex(index uint32) interface{} {
	node := l.head.next
	for i := uint32(1); i < l.len>>1; i++ {
		if i == index {
			node.prev.next = node.next
			node.next.prev = node.prev
			l.len--
			l.getLen++
			return node.value
		}
		node = node.next
	}
	return nil
}

//rPopByIndex 如果位置在右半边 则从后往前遍历寻找取出
func (l *GoList) rPopByIndex(index uint32) interface{} {
	node := l.tail.prev
	for i := l.len - 2; i >= (l.len >> 1); i-- {
		if i == index {
			node.prev.next = node.next
			node.next.prev = node.prev
			l.len--
			l.getLen++
			return node.value
		}
		node = node.prev
	}
	return nil
}

//RPop 从链表尾部取数据
func (l *GoList) RPop() interface{} {
	if l.len == 0 {
		return nil
	}

	node := l.tail
	l.tail = node.prev
	if l.tail == nil {
		l.head = nil
	} else {
		l.tail.next = nil
	}
	l.len--
	l.getLen++

	return node.value
}

//LPop 从链表头部取数据
func (l *GoList) LPop() interface{} {
	if l.len == 0 {
		return nil
	}

	node := l.head
	l.head = node.next
	if l.head == nil {
		l.tail = nil
	} else {
		l.head.prev = nil
	}
	l.len--
	l.getLen++

	return node.value
}

//Match  匹配interface{} value值相同的节点  返回匹配的节点ListNode
//key 为用户传进来的数值 这边原样传出去
//value 为节点存放的数值
//返回节点在链表的所在的位置索引index
func (l *GoList) Match(key interface{}, fn func(key, value interface{}) bool) (uint32, bool) {
	node := l.head
	for i := uint32(0); i < l.len; i++ {
		if fn(key, node.value) {
			return i, true
		}
		node = node.next
	}
	return 0, false
}

//MatchAndRemove 匹配value值相同的interface{}:为节点的存的值 并且返回删除的节点
//key 为用户传进来的数值 这边原样传出去
//value 为节点存放的数值
func (l *GoList) MatchAndRemove(key interface{}, fn func(key, value interface{}) bool) bool {
	node := l.head
	for i := uint32(0); i < l.len; i++ {
		if fn(key, node.value) {
			//删除节点
			l.Pop(i)
			return true
		}
		node = node.next
	}
	return false
}
