package classes

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

const Version = "0.2.7"

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

	classes.Classes["default"] = nil
	classes.SetThreshold("default", "ham", HAM_THRESHOLD)
	classes.SetThreshold("default", "possible", POSSIBLE_THRESHOLD)
	classes.SetThreshold("default", "probable", PROBABLE_THRESHOLD)
	classes.SetThreshold("default", "spam", MAX_THRESHOLD)

	return &classes, nil
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
	for addr, classes := range configClasses {
		sort.Sort(ByScore(classes))
		c.Classes[addr] = classes
	}
	return nil
}

func (c *SpamClasses) Write(filename string) error {
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
	classes, ok := c.Classes[address]
	if ok {
		return classes
	}
	classes, ok = c.Classes["default"]
	if ok {
		return classes
	}
	return []SpamClass{}
}

func (c *SpamClasses) GetClass(addresses []string, score float32) string {
	classes, ok := c.Classes["default"]
	if !ok {
		return ""
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
	_, ok := c.Classes[address]
	if !ok {
		c.Classes[address] = make([]SpamClass, 0)
	}
	var exists bool
	for i, class := range c.Classes[address] {
		if class.Name == name {
			c.Classes[address][i].Score = threshold
			exists = true
			break
		}
	}
	if !exists {
		c.Classes[address] = append(c.Classes[address], SpamClass{Name: name, Score: threshold})
	}
	sort.Sort(ByScore(c.Classes[address]))
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
