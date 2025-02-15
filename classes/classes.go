package classes

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

const Version = "1.0.1"

const HAM_THRESHOLD = float32(0)
const POSSIBLE_THRESHOLD = float32(3)
const PROBABLE_THRESHOLD = float32(10)
const MAX_THRESHOLD = float32(999)
const MAX_NAME = "spam"
const DEFAULT_NAME = "default"

type SpamClass struct {
	Name  string  `json:"name"`
	Score float32 `json:"score"`
}

type SpamClasses struct {
	Classes map[string][]SpamClass
}

var DefaultClasses = []SpamClass{
	{"ham", HAM_THRESHOLD},
	{"possible", POSSIBLE_THRESHOLD},
	{"probable", PROBABLE_THRESHOLD},
	{MAX_NAME, MAX_THRESHOLD},
}

type ByScore []SpamClass

func (a ByScore) Len() int {
	return len(a)
}

func (a ByScore) Less(i, j int) bool {
	return a[i].Score < a[j].Score
}

func (a ByScore) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func dupClasses(classes []SpamClass) []SpamClass {
	ret := []SpamClass{}
	for _, class := range classes {
		ret = append(ret, SpamClass{class.Name, class.Score})
	}
	return ret
}

func New(filename string) (*SpamClasses, error) {
	classes := SpamClasses{
		Classes: make(map[string][]SpamClass, 0),
	}

	if filename != "" {
		err := classes.Read(filename)
		if err != nil {
			return nil, err
		}
	}

	classes.Classes[DEFAULT_NAME] = DefaultClasses
	return &classes, nil
}

// validate all classes
func (c *SpamClasses) validateAll() {
	for address := range c.Classes {
		c.validate(address)
	}
}

// fix out of order or missing spam threshold
func (c *SpamClasses) validate(address string) {
	classes, ok := c.Classes[address]
	if !ok {
		return
	}
	nameMap := map[string]bool{MAX_NAME: true}
	scoreMap := map[float32]bool{MAX_THRESHOLD: true}
	valid := []SpamClass{{MAX_NAME, MAX_THRESHOLD}}

	for _, class := range classes {
		_, nameExists := nameMap[class.Name]
		if nameExists {
			continue
		}
		_, scoreExists := scoreMap[class.Score]
		if scoreExists {
			continue
		}
		nameMap[class.Name] = true
		scoreMap[class.Score] = true
		valid = append(valid, SpamClass{class.Name, class.Score})
	}
	sort.Sort(ByScore(valid))
	c.Classes[address] = valid
}

func (c *SpamClasses) Read(filename string) error {

	if filename == "" {
		return nil
	}
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return nil
	}
	configBytes, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed reading %s: %v", filename, err)
	}
	configClasses := map[string][]SpamClass{}
	err = json.Unmarshal(configBytes, &configClasses)
	if err != nil {
		return fmt.Errorf("failed parsing %s: %v", filename, err)
	}
	c.Classes = configClasses
	_, ok := c.Classes[DEFAULT_NAME]
	if !ok {
		c.Classes[DEFAULT_NAME] = dupClasses(DefaultClasses)
	}

	c.validateAll()

	return nil
}

func (c *SpamClasses) Write(filename string) error {
	c.validateAll()
	data, err := json.MarshalIndent(&c.Classes, "", "  ")
	if err != nil {
		return fmt.Errorf("failed marshalling: %v", err)
	}
	err = os.WriteFile(filename, data, 0660)
	if err != nil {
		return fmt.Errorf("failed writing %s: %v", filename, err)
	}
	return nil
}

// return slice of SpamClass for address or default;  always returns a list
func (c *SpamClasses) GetClasses(address string) []SpamClass {
	var classes []SpamClass
	var ok bool
	classes, ok = c.Classes[address]
	if ok {
		return classes
	}
	classes, ok = c.Classes[DEFAULT_NAME]
	if !ok {
		classes = DefaultClasses
	}
	c.Classes[address] = dupClasses(classes)
	c.validate(address)
	return c.Classes[address]
}

func (c *SpamClasses) SetClasses(address string, classes []SpamClass) []SpamClass {
	c.Classes[address] = dupClasses(classes)
	c.validate(address)
	return c.Classes[address]
}

func (c *SpamClasses) GetClass(addresses []string, score float32) string {
	classes := []SpamClass{}
	for _, address := range addresses {
		var ok bool
		classes, ok = c.Classes[address]
		if ok {
			break
		}
	}
	if len(classes) == 0 {
		classes = c.GetClasses(DEFAULT_NAME)
	}
	var result string
	for _, class := range classes {
		result = class.Name
		if score < class.Score {
			break
		}
	}
	return result
}

func (c *SpamClasses) SetThreshold(address, name string, threshold float32) {
	if name == MAX_NAME {
		threshold = MAX_THRESHOLD
	}
	classes := c.GetClasses(address)
	var exists bool
	for i, class := range classes {
		if class.Name == name {
			classes[i].Score = threshold
			exists = true
			break
		}
	}
	if !exists {
		classes = append(classes, SpamClass{Name: name, Score: threshold})
	}
	c.Classes[address] = classes
	c.validate(address)
}

func (c *SpamClasses) GetThreshold(address, name string) (float32, bool) {
	classes := c.GetClasses(address)
	for _, class := range classes {
		if class.Name == name {
			return class.Score, true
		}
	}
	return 0.0, false
}

func (c *SpamClasses) DeleteClasses(address string) {
	delete(c.Classes, address)
}

func (c *SpamClasses) DeleteClass(address, name string) {
	if name == MAX_NAME {
		return
	}
	classes := c.GetClasses(address)
	for i, class := range classes {
		if class.Name == name {
			c.Classes[address] = append(c.Classes[address][:i], c.Classes[address][i+1:]...)
			if len(c.Classes[address]) == 0 {
				c.DeleteClasses(address)
			} else {
				c.validate(address)
			}
			return
		}
	}
}

func (c *SpamClasses) Usernames() []string {
	users := []string{}
	for key, _ := range c.Classes {
		users = append(users, key)
	}
	return users
}
