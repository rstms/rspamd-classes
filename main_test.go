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

func TestDeleteClasses(t *testing.T) {
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
	c.SetThreshold("bravo", "higher", 999)
	require.Equal(t, c.Classes["bravo"][0], classes.SpamClass{"zero", 0})
	require.Equal(t, c.Classes["bravo"][1], classes.SpamClass{"low", 1})
	require.Equal(t, c.Classes["bravo"][2], classes.SpamClass{"medium", 5})
	require.Equal(t, c.Classes["bravo"][3], classes.SpamClass{"high", 10})
	require.Equal(t, c.Classes["bravo"][4], classes.SpamClass{"higher", 999})
	dump(t, c.Classes["bravo"])
}
