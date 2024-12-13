package classes

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

const Version = "0.2.9"

const HAM_THRESHOLD = 0.0
const POSSIBLE_THRESHOLD = 3.0
const PROBABLE_THRESHOLD = 10.0
const MAX_THRESHOLD = 999.0

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
	{"spam", MAX_THRESHOLD},
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

	classes.Classes["default"] = DefaultClasses
	return &classes, nil
}

// fix out of order or missing spam threshold
func (c *SpamClasses) validate(address string) {
	classes, ok := c.Classes[address]
	if !ok {
		return
	}
	classMap := map[string]float32{}
	for _, class := range classes {
		if class.Score < MAX_THRESHOLD {
			classMap[class.Name] = class.Score
		}
	}
	classMap["spam"] = MAX_THRESHOLD
	valid := []SpamClass{}
	for name, score := range classMap {
		valid = append(valid, SpamClass{name, score})
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
	for addr := range c.Classes {
		c.validate(addr)
	}
	return nil
}

func (c *SpamClasses) Write(filename string) error {
	for address := range c.Classes {
		c.validate(address)
	}
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
	_, ok := c.Classes[address]
	if ok {
		c.validate(address)
		return c.Classes[address]
	}
	_, ok = c.Classes["default"]
	if ok {
		c.validate("default")
		return c.Classes["default"]
	}
	return DefaultClasses
}

func (c *SpamClasses) GetClass(addresses []string, score float32) string {
	classes, ok := c.Classes["default"]
	if !ok {
		classes = DefaultClasses
	}
	for _, address := range addresses {
		addrClasses, ok := c.Classes[address]
		if ok {
			classes = addrClasses
			break
		}
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
	classes, ok := c.Classes[address]
	if !ok {
		classes = make([]SpamClass, 0)
	}
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
	classes, ok := c.Classes[address]
	if ok {
		for _, class := range classes {
			if class.Name == name {
				return class.Score, true
			}
		}
	}
	return 0.0, false
}

func (c *SpamClasses) DeleteClasses(address string) {
	delete(c.Classes, address)
}

func (c *SpamClasses) DeleteClass(address, name string) {
	_, ok := c.Classes[address]
	if ok {
		for i, class := range c.Classes[address] {
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
}

func (c *SpamClasses) Usernames() []string {
	users := []string{}
	for key, _ := range c.Classes {
		users = append(users, key)
	}
	return users
}
