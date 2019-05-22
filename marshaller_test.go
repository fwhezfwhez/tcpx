package tcpx

import (
	"encoding/xml"
	"fmt"
	"github.com/fwhezfwhez/errorx"
	"testing"
)

func TestJsonMarshaller(t *testing.T) {
	fmt.Println(JsonMarshaller{}.MarshalName())
	fmt.Println(JsonMarshaller{}.Marshal("123"))
	var str string
	fmt.Println(JsonMarshaller{}.Unmarshal([]byte(`"123"`), &str), str)
}

func TestXmlMarshaller(t *testing.T) {
	fmt.Println(XmlMarshaller{}.MarshalName() == "xml")
	type User struct {
		XMLName  xml.Name `xml:"xml"`
		Username string   `xml:"username"`
	}
	buf, e := XmlMarshaller{}.Marshal(User{Username: "Ft"})
	if e != nil {
		fmt.Println(e.Error())
		t.Fail()
		return
	}
	var user User
	e = XmlMarshaller{}.Unmarshal(buf, &user)
	fmt.Println(user)
	if e != nil {
		fmt.Println(e.Error())
		t.Fail()
		return
	}
}

func TestTOMLYAMLMarshaller(t *testing.T) {
	fmt.Println(TomlMarshaller{}.MarshalName() == "toml")
	fmt.Println(YamlMarshaller{}.MarshalName() == "yaml")
	type User struct {
		Username string `toml:"username" yaml:"username"`
	}
	buf, e := TomlMarshaller{}.Marshal(User{"ft"})
	fmt.Println(string(buf))
	if e != nil {
		fmt.Println(e.Error())
		t.Fail()
		return
	}
	var user User
	e = TomlMarshaller{}.Unmarshal(buf, &user)
	fmt.Println(user)
	if e != nil {
		fmt.Println(e.Error())
		t.Fail()
		return
	}

	buf, e = YamlMarshaller{}.Marshal(User{"ft"})
	if e != nil {
		fmt.Println(e.Error())
		t.Fail()
		return
	}
	fmt.Println(string(buf))
	var user2 User
	e = YamlMarshaller{}.Unmarshal(buf, &user2)
	fmt.Println(user2)
	if e != nil {
		fmt.Println(e.Error())
		t.Fail()
		return
	}
}

func TestGetMarshallerByMarshalName(t *testing.T) {
	jsonMarshaller, e := GetMarshallerByMarshalName("json")
	if e != nil {
		fmt.Println(errorx.Wrap(e).Error())
		t.Fail()
		return
	}
	xmlMarshaller, e := GetMarshallerByMarshalName("xml")
	if e != nil {
		fmt.Println(errorx.Wrap(e).Error())
		t.Fail()
		return
	}
	tomlMarshaller, e := GetMarshallerByMarshalName("toml")
	if e != nil {
		fmt.Println(errorx.Wrap(e).Error())
		t.Fail()
		return
	}
	yamlMarshaller, e := GetMarshallerByMarshalName("yaml")
	if e != nil {
		fmt.Println(errorx.Wrap(e).Error())
		t.Fail()
		return
	}
	_, e = GetMarshallerByMarshalName("xxx")
	if e == nil {
		fmt.Println("should throw err but receive nothing")
		t.Fail()
		return
	}
	fmt.Println(
		jsonMarshaller.MarshalName(),
		xmlMarshaller.MarshalName(),
		tomlMarshaller.MarshalName(),
		yamlMarshaller.MarshalName(),
	)

}
