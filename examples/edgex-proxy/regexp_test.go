package main

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

func TestMethod(t *testing.T) {
	topic := "$iothub/methods/POST/Values/?$rid=1"
	rid := topic[len(topic)-1 : len(topic)]
	fmt.Println("rid: " + rid)
	m, err := regexp.Compile("POST/(.*?)/")
	if err != nil {
		fmt.Println(err.Error())
	}

	methodName := m.FindString(topic)
	fmt.Println("methodName: " + methodName)
	methodName = strings.Replace(methodName, "POST", "", 1)
	methodName = strings.Replace(methodName, "/", "", -1)

	fmt.Println("methodName: " + methodName)
}
