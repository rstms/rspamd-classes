package main

import (
	"encoding/json"
	"github.com/rstms/rspamd-classes/classes"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
)

func TestVersion(t *testing.T) {
	config, err := classes.New("")
	require.Nil(t, err)
	require.NotNil(t, config)
}

func dump(t *testing.T, d interface{}) {
	data, err := json.MarshalIndent(d, "", "  ")
	require.Nil(t, err)
	log.Println(string(data))
}

func TestEdits(t *testing.T) {
	config, err := classes.New("")
	require.Nil(t, err)
	config.SetThreshold("test", "low", 1)
	config.SetThreshold("test", "medium", 5)
	config.SetThreshold("test", "high", 10)
	dump(t, config)
	err = config.Write("testdata/classes.json")
	require.Nil(t, err)

	readback, err := classes.New("testdata/classes.json")
	require.Nil(t, err)
	dump(t, readback)

	require.Equal(t, config, readback)
}

func initClasses(t *testing.T) *classes.SpamClasses {
	config, err := classes.New("")
	require.Nil(t, err)
	config.SetThreshold("alpha", "low", 1)
	config.SetThreshold("alpha", "medium", 5)
	config.SetThreshold("alpha", "high", 10)
	config.SetThreshold("bravo", "low", 1)
	config.SetThreshold("bravo", "medium", 5)
	config.SetThreshold("bravo", "high", 10)
	return config
}

func TestDeleteAllClasses(t *testing.T) {
	config := initClasses(t)
	config.Write("testdata/alpha_bravo.json")
	require.Len(t, config.Classes, 3)
	config.DeleteClasses("alpha")
	config.Write("testdata/bravo.json")
	require.Len(t, config.Classes, 2)

	read_alpha_bravo, err := classes.New("testdata/alpha_bravo.json")
	require.Nil(t, err)
	require.Len(t, read_alpha_bravo.Classes, 3)
	users := read_alpha_bravo.Usernames()
	require.Len(t, users, 3)
	require.Contains(t, users, "alpha")
	require.Contains(t, users, "bravo")
	require.Contains(t, users, "default")

	read_bravo, err := classes.New("testdata/bravo.json")
	require.Nil(t, err)
	require.Len(t, read_bravo.Classes, 2)
	users = read_bravo.Usernames()
	require.Len(t, users, 2)
	require.Contains(t, users, "bravo")
	require.Contains(t, users, "default")
}

func TestChangeClassThreshold(t *testing.T) {
	c := initClasses(t)
	require.Equal(t, c.Classes["alpha"][0], classes.SpamClass{"low", 1})
	require.Equal(t, c.Classes["alpha"][1], classes.SpamClass{"medium", 5})
	require.Equal(t, c.Classes["alpha"][2], classes.SpamClass{"high", 10})

	c.SetThreshold("alpha", "medium", 5.5)

	dump(t, c)
	require.Equal(t, c.Classes["alpha"][1], classes.SpamClass{"medium", 5.5})
}

func TestGetClass(t *testing.T) {
	c := initClasses(t)
	require.Equal(t, "low", c.GetClass([]string{"bravo"}, -1))
	c.SetThreshold("bravo", "low", -2)
	dump(t, c.Classes["bravo"])
	require.Equal(t, "medium", c.GetClass([]string{"bravo"}, -1))
	require.Equal(t, "medium", c.GetClass([]string{"bravo"}, 0))
	dump(t, c.Classes["bravo"])
	c.SetThreshold("bravo", "new", .5)
	dump(t, c.Classes["bravo"])
	require.Equal(t, "new", c.GetClass([]string{"bravo"}, 0))
}

func TestInsertSort(t *testing.T) {
	c := initClasses(t)
	c.SetThreshold("bravo", "zero", 0)
	c.SetThreshold("bravo", "higher", 100)
	require.Equal(t, c.Classes["bravo"][0], classes.SpamClass{"zero", 0})
	require.Equal(t, c.Classes["bravo"][1], classes.SpamClass{"low", 1})
	require.Equal(t, c.Classes["bravo"][2], classes.SpamClass{"medium", 5})
	require.Equal(t, c.Classes["bravo"][3], classes.SpamClass{"high", 10})
	require.Equal(t, c.Classes["bravo"][4], classes.SpamClass{"higher", 100})
	require.Equal(t, c.Classes["bravo"][5], classes.SpamClass{"spam", 999})
	dump(t, c.Classes["bravo"])
}
func TestDeleteClass(t *testing.T) {
	c, err := classes.New("")
	require.Nil(t, err)
	c.SetThreshold("test", "one", 1)
	c.SetThreshold("test", "two", 2)
	c.SetThreshold("test", "three", 3)

	list, ok := c.Classes["test"]
	require.True(t, ok)
	require.Len(t, list, 4)
	require.Equal(t, list[0], classes.SpamClass{"one", 1})
	require.Equal(t, list[1], classes.SpamClass{"two", 2})
	require.Equal(t, list[2], classes.SpamClass{"three", 3})
	require.Equal(t, list[3], classes.SpamClass{"spam", 999})

	c.DeleteClass("test", "two")

	list, ok = c.Classes["test"]
	require.True(t, ok)
	require.Len(t, list, 3)
	require.Equal(t, list[0], classes.SpamClass{"one", 1})
	require.Equal(t, list[1], classes.SpamClass{"three", 3})
	require.Equal(t, list[2], classes.SpamClass{"spam", 999})

	c.DeleteClass("test", "one")
	list, ok = c.Classes["test"]
	require.True(t, ok)
	require.Len(t, list, 2)
	require.Equal(t, list[0], classes.SpamClass{"three", 3})
	require.Equal(t, list[1], classes.SpamClass{"spam", 999})

	c.DeleteClass("test", "fnord")
	list, ok = c.Classes["test"]
	require.True(t, ok)
	require.Len(t, list, 2)
	require.Equal(t, list[0], classes.SpamClass{"three", 3})
	require.Equal(t, list[1], classes.SpamClass{"spam", 999})

	c.DeleteClass("test", "three")
	list, ok = c.Classes["test"]
	require.True(t, ok)
	require.Equal(t, list[0], classes.SpamClass{"spam", 999})

	c.DeleteClass("test", "spam")
	list, ok = c.Classes["test"]
	require.False(t, ok)

}

func TestGetClasses(t *testing.T) {
	c := initClasses(t)

	alist := c.GetClasses("alpha")
	require.Len(t, alist, 4)
	require.Equal(t, alist[0], classes.SpamClass{"low", 1})
	require.Equal(t, alist[1], classes.SpamClass{"medium", 5})
	require.Equal(t, alist[2], classes.SpamClass{"high", 10})
	require.Equal(t, alist[3], classes.SpamClass{"spam", 999})

	dlist := c.GetClasses("default")
	require.Len(t, dlist, 4)
	require.Equal(t, dlist[0], classes.SpamClass{"ham", 0})
	require.Equal(t, dlist[1], classes.SpamClass{"possible", 3})
	require.Equal(t, dlist[2], classes.SpamClass{"probable", 10})
	require.Equal(t, dlist[3], classes.SpamClass{"spam", 999})

	nlist := c.GetClasses("nonexistent")
	require.Len(t, nlist, 4)
	require.Equal(t, dlist, nlist)
}
