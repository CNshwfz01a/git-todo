package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

type todoList struct {
	Id      int    `json:"序号"`
	Content string `json:"内容"`
	IsDone  bool   `json:"是否完成"`
}

// 获取当前git分支名
func getCurrentBranch() string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(out.String())
}

// 根据开始和结尾字符获取字符串内容
func getInputByDelimiters(start, end, content string) string {
	startIndex := strings.Index(content, start)
	endIndex := strings.Index(content, end)
	if startIndex == -1 || endIndex == -1 || startIndex >= endIndex {
		return ""
	}
	return content[startIndex+len(start) : endIndex]
}

func jsonFileToStruct(path string) (todoList, string) {
	var list todoList
	jsonFile, err := os.Open(path)
	if err != nil {
		fmt.Println("error opening json file")
		return list, "error opening json file"
	}
	defer jsonFile.Close()
	decoder := json.NewDecoder(jsonFile)
	for {
		err := decoder.Decode(&list)
		if err == io.EOF {
			break
		}
	}
	return list, "success"
}

func readFromJsonFile(path string) (int, string, []todoList) {
	// fmt.Println("正在读取代办文件:", path)
	var list []todoList
	//判断文件是否存在
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		// fmt.Println("file not exist")
		//返回空结构和错误
		// fmt.Println("当前分支不存在代办文件")
		return 1, "当前分支不存在代办", list
	}
	jsonFile, err := os.Open(path)
	if err != nil {
		// fmt.Println("error opening json file")
		//返回空结构和错误
		return 2, "error opening json file", list
	}
	defer jsonFile.Close()

	jsonData, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		// fmt.Println("error reading json file")
		return 2, "error reading json file", list
	}
	json.Unmarshal(jsonData, &list)
	// fmt.Println(list)
	// fmt.Println("代办文件读取成功")
	return 3, "success", list
}

func writeStructDataToJsonFile(list todoList) {
	fmt.Println("正在增加代办:", list.Content)
	branch := getCurrentBranch()
	//判断是否存在代办文件
	filePath := fmt.Sprintf("./todo-list-%s.json", branch)
	var data []todoList
	var message string
	var res int
	res, message, data = readFromJsonFile(filePath)
	//如果不存在文件则创建
	if res == 1 {
		fmt.Println(message)
		fmt.Println("正在创建代办文件:", filePath)
		res, err := os.Create(filePath)
		if nil != err {
			panic(err)
		}
		defer res.Close()
	} else if res == 2 {
		fmt.Println("读取代办文件失败:", message)
		return
	}
	// encoder := json.NewEncoder(os.Stdout)
	//获取当前最大id
	maxId := 0
	for _, item := range data {
		if item.Id > maxId {
			maxId = item.Id
		}
	}
	list.Id = maxId + 1
	data = append(data, list)
	// encoder.Encode(data)
	//写入文件
	// fmt.Println("正在写入代办文件:", filePath)
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println("error marshalling json data")
		return
	}
	//写入文件
	err = ioutil.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		fmt.Println("error writing json data to file")
		return
	}
	fmt.Println("代办添加成功")
}

func list() {
	var data []todoList
	var message string
	var res int
	branch := getCurrentBranch()
	filePath := fmt.Sprintf("./todo-list-%s.json", branch)
	//读取文件
	res, message, data = readFromJsonFile(filePath)
	if res == 1 || res == 2 {
		fmt.Println("读取代办文件失败:", message)
		// return
	}
	// fmt.Println("代办列表如下:", data)
	//循环打印
	for _, item := range data {
		fmt.Printf("%d %s 完成:%t\n", item.Id, item.Content, item.IsDone)
	}
}

func add(Content string) {
	var ContentData todoList
	ContentData.Content = Content
	ContentData.IsDone = false
	writeStructDataToJsonFile(ContentData)
}

func done(Id int) {
	branch := getCurrentBranch()
	filePath := fmt.Sprintf("./todo-list-%s.json", branch)
	var data []todoList
	var message string
	var res int
	res, message, data = readFromJsonFile(filePath)
	if res == 1 || res == 2 {
		fmt.Println("读取代办文件失败:", message)
		return
	}
	for i, item := range data {
		if item.Id == Id {
			data[i].IsDone = true
			break
		}
	}
	//写入文件
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println("error marshalling json data")
		return
	}
	err = ioutil.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		fmt.Println("error writing json data to file")
		return
	}
}

func delete(Id int) {
	branch := getCurrentBranch()
	filePath := fmt.Sprintf("./todo-list-%s.json", branch)
	var data []todoList
	var message string
	var res int
	res, message, data = readFromJsonFile(filePath)
	if res == 1 || res == 2 {
		fmt.Println("读取代办文件失败:", message)
		return
	}
	for i, item := range data {
		if item.Id == Id {
			data = append(data[:i], data[i+1:]...)
			break
		}
	}
	//重新排序数组
	for i := range data {
		data[i].Id = i + 1
	}
	//写入文件
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println("error marshalling json data")
		return
	}
	err = ioutil.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		fmt.Println("error writing json data to file")
		return
	}
	fmt.Println("序号", Id, "代办删除成功")
}

func main() {
	//判断命令行参数
	var input = os.Args
	if len(input) < 2 {
		fmt.Println("使用git todo help命令获取操作方式")
		return
	}
	// fmt.Println("命令参数为:", input[1])
	//根据git todo 命令执行动作
	switch input[1] {
	case "help":
		fmt.Println("git todo list - 列出所有代办")
		fmt.Println("git todo add <内容> - 增加代办")
		fmt.Println("git todo done [1] - 完成1代办")
		fmt.Println("git todo delete [1] - 删除1代办")
	case "list":
		list()
	case "add":
		//判断是否有内容参数
		if len(input) < 3 {
			fmt.Println("请提供代办内容")
			return
		}
		//获取"<>"中的内容
		content := getInputByDelimiters("<", ">", input[2])
		if content == "" {
			fmt.Println("请提供正确格式的代办内容，使用<>包裹内容")
			return
		}
		add(content)
	case "done":
		//获取"[]"中的内容并转换为int
		idStr := getInputByDelimiters("[", "]", input[2])
		var id int
		fmt.Sscanf(idStr, "%d", &id)
		done(id)
	case "delete":
		//获取"[]"中的内容并转换为int
		idStr := getInputByDelimiters("[", "]", input[2])
		var id int
		fmt.Sscanf(idStr, "%d", &id)
		delete(id)
	}
}
