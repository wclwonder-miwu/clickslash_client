package main

import (
	"clickslash/Im"
	"fmt"
	"log"
	"testing"

	"github.com/golang/protobuf/proto"
)

func TestMain(t *testing.T) {
	testGrammar()
}

func testGrammar() {
	//checkAlone("1111111")
	fmt.Println(len("寒"))
}

func quickSort2(values []int, left, right int) {
	i := left
	j := right

	if i < j { //如果已经到结尾，只剩一个数的等
		temp := values[left]
		for i < j {
			//从右往左找到比temp小的放到temp，空出j
			for j > i && values[j] >= temp {
				j--
			}
			if j > i {
				values[i] = values[j]
				i++
			}

			//从左往右找到比temp大的放到j，空出i位置
			for i < j && values[i] <= temp {
				i++
			}
			if i < j {
				values[j] = values[i]
				j--
			}
		}

		//i==j时结束一轮
		values[i] = temp
		quickSort2(values, left, i-1)
		quickSort2(values, i+1, right)
	}
}

func quickSort(values []int, left, right int) {
	temp := values[left]
	p := left
	i, j := left, right

	for i <= j {
		for j >= p && values[j] >= temp {
			j--
		}
		if j >= p {
			values[p] = values[j]
			p = j
		}

		if values[i] <= temp && i <= p {
			i++
		}
		if i <= p {
			values[p] = values[i]
			p = i
		}
	}
	values[p] = temp
	fmt.Println(p)
	fmt.Println(values)
	if p-left > 1 {
		quickSort(values, left, p-1)
	}
	if right-p > 1 {
		quickSort(values, p+1, right)
	}
}

func checkAlone(str string) []rune {
	strRune := []rune(str)
	retRune := []rune{}

	for i, chi := range strRune {
		find := false
		for j, chj := range strRune {
			if i != j && chi == chj {
				find = true
			}
		}

		if find == false {
			retRune = append(retRune, chi)
		}
	}
	fmt.Println(string(retRune))
	return retRune
}

func testIm() {
	// 创建一个消息
	test := &Im.Helloworld{
		// 使用辅助函数设置域的值
		//Str: "hello!" ,
		//  Id:  321,
		Opt: 1234,
	}

	test.Id = 3244

	// 进行编码
	data, err := proto.Marshal(test)
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}

	// 进行解码
	newTest := &Im.Helloworld{}
	err = proto.Unmarshal(data, newTest)
	if err != nil {
		log.Fatal("unmarshaling error: ", err)
	}

	log.Printf("id:%d;opt:%d;str:%s;", newTest.Id, newTest.Opt, newTest.Str)

	// 测试结果
	if test.String() != newTest.String() {
		log.Fatalf("data mismatch %q != %q", test.String(), newTest.String())
	}
}
