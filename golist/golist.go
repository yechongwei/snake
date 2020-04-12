package golist

import (
	"fmt"
)

//IList 接口类
type IList interface {
	Len() int
	IsEmpty() bool
	Clear()
	ListGetIterator() IIterator
	Push(index int, data interface{}) bool
	RPush(data interface{}) bool
	LPush(data interface{}) bool
	Pop(index int) *ListNode
	RPop() *ListNode
	LPop() *ListNode
	//匹配value值相同的 interface{}:为节点的存的值 返回匹配的节点
	Match(key interface{}, fn func(key, Value interface{}) bool) *ListNode
	//匹配value值相同的interface{}:为节点的存的值 并且返回是否找到并且删除
	MatchAndRemove(key interface{}, fn func(key, Value interface{}) bool) *ListNode
}

//ListNode 链表节点
//Value 节点存放的数据
//Prev 上一个节点
//Next  下一个节点
type ListNode struct {
	Value interface{}
	Prev  *ListNode
	Next  *ListNode
}

//IIterator 迭代器接口
type IIterator interface {
	//Next 迭代器获取下一个节点
	Next() *ListNode
}

//ListIterator 链表迭代器 配合Next使用
type ListIterator struct {
	next *ListNode
}

//Next 获取下个节点的数据
func (lter *ListIterator) Next() *ListNode {
	node := lter.next
	if node != nil {
		lter.next = node.Next
	}
	return node
}

//GoList 链表  多线程不安全 外层需要加锁保护
//head 链表头节点
//tail 链表尾节点
//len 链表长度
type GoList struct {
	head   *ListNode
	tail   *ListNode
	len    int
	getLen uint64
	putLen uint64
}

func newNode(data interface{}, prev, next *ListNode) *ListNode {
	return &ListNode{
		Value: data,
		Prev:  prev,
		Next:  next,
	}
}

func clearNode(node *ListNode) {
	if node != nil {
		node.Value = nil
		node.Prev, node.Next = nil, nil
		node = nil
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
func (l *GoList) Len() int {
	return l.len
}

//IsEmpty 链表是否为空
func (l *GoList) IsEmpty() bool {
	return l.len == 0
}

//ListGetIterator 生成迭代器 配合Next 使用
func (l *GoList) ListGetIterator() IIterator {
	return &ListIterator{next: l.head}
}

//Push 往链表固定位置存放数据
//存放成功返回true  失败返回false
func (l *GoList) Push(index int, data interface{}) bool {
	if index < 0 {
		index = l.len + index
	}

	if index <= 0 {
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
func (l *GoList) lPushByIndex(index int, data interface{}) bool {
	nodeTmp := l.head
	for i := 1; index > 1 && i <= l.len>>1; i++ {
		index--
		nodeTmp = nodeTmp.Next
	}

	if node := newNode(data, nodeTmp, nodeTmp.Next); node != nil {
		nodeTmp.Next.Prev = node
		nodeTmp.Next = node
		l.len++
		l.putLen++
		return true
	}

	return false
}

//rPushByIndex 如果插入位置在右半边 则从尾往头遍历寻找插入
func (l *GoList) rPushByIndex(index int, data interface{}) bool {
	nodeTmp := l.tail
	for i := l.len - 1; index < l.len-1 && i > l.len>>1; i-- {
		index++
		nodeTmp = nodeTmp.Prev
	}

	if node := newNode(data, nodeTmp.Prev, nodeTmp); node != nil {
		nodeTmp.Prev.Next = node
		nodeTmp.Prev = node
		l.len++
		l.putLen++
		return true
	}
	return false
}

//RPush  往链表尾部后插入
func (l *GoList) RPush(data interface{}) bool {
	if node := newNode(data, l.tail, nil); node != nil {
		if l.tail == nil {
			l.head = node
		} else {
			l.tail.Next = node
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
			l.head.Prev = node
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
func (l *GoList) Pop(index int) *ListNode {
	if l.len == 0 {
		return nil
	}

	if index < 0 {
		index = l.len + index
	}

	if index <= 0 {
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
func (l *GoList) lPopByIndex(index int) *ListNode {
	node := l.head.Next
	for i := 1; index > 1 && i < l.len>>1; i++ {
		index--
		node = node.Next
	}

	node.Prev.Next = node.Next
	node.Next.Prev = node.Prev
	l.len--
	l.getLen++
	return node
}

//rPopByIndex 如果位置在右半边 则从后往前遍历寻找取出
func (l *GoList) rPopByIndex(index int) *ListNode {
	node := l.tail.Prev
	for i := l.len - 2; index < l.len-2 && i >= (l.len>>1); i-- {
		index++
		node = node.Prev
	}

	node.Prev.Next = node.Next
	node.Next.Prev = node.Prev
	l.len--
	l.getLen++
	return node
}

//RPop 从链表尾部取数据
func (l *GoList) RPop() *ListNode {
	if l.len == 0 {
		return nil
	}

	node := l.tail
	l.tail = node.Prev
	if l.tail == nil {
		l.head = nil
	} else {
		l.tail.Next = nil
	}
	l.len--
	l.getLen++

	return node
}

//LPop 从链表头部取数据
func (l *GoList) LPop() *ListNode {
	if l.len == 0 {
		return nil
	}

	node := l.head
	l.head = node.Next
	if l.head == nil {
		l.tail = nil
	} else {
		l.head.Prev = nil
	}
	l.len--
	l.getLen++

	return node
}

//Match  匹配interface{} value值相同的节点  返回匹配的节点ListNode
//key 为用户传进来的数值 这边原样传出去
//Value 为节点存放的数值
func (l *GoList) Match(key interface{}, fn func(key, Value interface{}) bool) *ListNode {
	node := l.head
	for i := 0; i < l.len; i++ {
		if fn(key, node.Value) {
			return node
		}
		node = node.Next
	}
	return nil
}

//MatchAndRemove 匹配value值相同的interface{}:为节点的存的值 并且返回删除的节点
//key 为用户传进来的数值 这边原样传出去
//Value 为节点存放的数值
func (l *GoList) MatchAndRemove(key interface{}, fn func(key, Value interface{}) bool) *ListNode {
	node := l.head
	for i := 0; i < l.len; i++ {
		if fn(key, node.Value) {
			//删除节点
			return l.Pop(i)
		}
		node = node.Next
	}
	return nil
}

//Clear 清除链表
func (l *GoList) Clear() {
	var current, next *ListNode = l.head, nil
	for i := 0; i < l.len; i++ {
		next = current.Next
		clearNode(current)
		current = next
	}
	l.head, l.tail = nil, nil
	l.getLen, l.putLen = 0, 0
	l.len = 0
}
