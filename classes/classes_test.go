package classes

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSpamLevelsNoFile(t *testing.T) {
	SpamClasses, err := New("")
	require.Nil(t, err)
	fmt.Printf("SpamClasses: %v\n", SpamClasses)
	require.Len(t, SpamClasses.Classes, 1)
	expected := []SpamClass{{"ham", 0.0}, {"possible", 3.0}, {"probable", 10.0}, {"spam", 999}}
	for key, value := range SpamClasses.Classes {
		require.Equal(t, key, "default")
		require.Equal(t, len(value), len(expected))
		for i, level := range value {
			require.Equal(t, level, expected[i])
		}
	}
}

func TestSpamLevelsConfigFile(t *testing.T) {
	SpamClasses, err := New("testdata/rspamd_classes.json")
	require.Nil(t, err)
	fmt.Printf("SpamClasses: %v\n", SpamClasses)
	require.Len(t, SpamClasses.Classes, 2)
	defaultLevels, ok := SpamClasses.Classes["default"]
	require.True(t, ok, "missing default classes")
	userLevels, ok := SpamClasses.Classes["username@example.org"]
	require.True(t, ok, "missing user classes")
	require.Equal(t, len(defaultLevels), len(userLevels))
	for i, level := range defaultLevels {
		require.Equal(t, level, userLevels[i])
	}
}

func TestWrite(t *testing.T) {
	SpamClasses, err := New("testdata/rspamd_classes.json")
	require.Nil(t, err)
	err = SpamClasses.Write("testdata/output.json")
	require.Nil(t, err)
	readback, err := New("testdata/output.json")
	require.Nil(t, err)
	for key, classes := range SpamClasses.Classes {
		rclasses, ok := readback.Classes[key]
		require.True(t, ok, "readback key not found: %s\n", key)
		fmt.Printf("key=%s classes: %v\n", key, classes)
		fmt.Printf("key=%s rclasses: %v\n", key, rclasses)
		for i, class := range classes {
			require.Equal(t, class, rclasses[i], "key=%s classes[%d] (%v) mismatches rclasses[%d] (%v)\n", key, i, class, i, rclasses[i])
		}
	}
}

func TestAdd(t *testing.T) {
	// init empty
	c, err := New("")
	require.Nil(t, err)

	// add low=1
	c.SetThreshold("user", "low", 1)
	list, ok := c.Classes["user"]
	require.True(t, ok)
	require.Len(t, list, 1)

	// change low=2
	c.SetThreshold("user", "low", 2)
	list, ok = c.Classes["user"]
	require.True(t, ok)
	require.Len(t, list, 1)

	// add med = 5
	c.SetThreshold("user", "medium", 5)
	list, ok = c.Classes["user"]
	require.True(t, ok)
	require.Len(t, list, 2)

	users := []string{"user"}
	class := c.GetClass(users, -1)
	require.Equal(t, class, "low")
	class = c.GetClass(users, 2)
	require.Equal(t, class, "medium")
	class = c.GetClass(users, 9)
	require.Equal(t, class, "medium")

	// add high=999
	c.SetThreshold("user", "high", 999)
	list, ok = c.Classes["user"]
	require.True(t, ok)
	require.Len(t, list, 3)

	class = c.GetClass(users, 9)
	require.Equal(t, class, "high")

	fmt.Printf("%v\n", c)
}

func TestSort(t *testing.T) {
	c, err := New("")
	require.Nil(t, err)
	c.SetThreshold("test", "nine", 9)
	c.SetThreshold("test", "five", 5)
	c.SetThreshold("test", "negone", -1)
	list, ok := c.Classes["test"]
	require.True(t, ok)
	require.Len(t, list, 3)
	require.Equal(t, list[0].Score, float32(-1))
	require.Equal(t, list[1].Score, float32(5))
	require.Equal(t, list[2].Score, float32(9))
}

