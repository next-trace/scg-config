package utils_test

import (
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/next-trace/scg-config/contract"
	"github.com/next-trace/scg-config/errors"
	"github.com/next-trace/scg-config/utils"
)

func TestEnvHelpers(t *testing.T) {
	t.Parallel()

	// NormalizeEnvKey
	require.Equal(t, "app.name", utils.NormalizeEnvKey("APP_NAME"))
	require.Equal(t, "a.b.c", utils.NormalizeEnvKey("A_B_C"))

	// NormalizePrefix
	require.Equal(t, "APP_", utils.NormalizePrefix("app"))
	require.Equal(t, "", utils.NormalizePrefix(""))

	// ShouldProcessEnv
	require.True(t, utils.ShouldProcessEnv("APP_NAME=ok", ""))
	require.True(t, utils.ShouldProcessEnv("APP_NAME=ok", "APP_"))
	require.False(t, utils.ShouldProcessEnv("OTHER_NAME=ok", "APP_"))

	// SplitEnv
	k, v := utils.SplitEnv("KEY=VALUE")
	require.Equal(t, "KEY", k)
	require.Equal(t, "VALUE", v)
	k, v = utils.SplitEnv("INVALID")
	require.Equal(t, "INVALID", k)
	require.Equal(t, "", v)

	// StripPrefix
	require.Equal(t, "NAME", utils.StripPrefix("APP_NAME", "APP_"))
	require.Equal(t, "APP_NAME", utils.StripPrefix("APP_NAME", ""))

	// IsSupportedConfigFile
	require.True(t, utils.IsSupportedConfigFile("file.yaml"))
	require.True(t, utils.IsSupportedConfigFile("file.yml"))
	require.True(t, utils.IsSupportedConfigFile("file.json"))
	require.False(t, utils.IsSupportedConfigFile("file.toml"))
	require.False(t, utils.IsSupportedConfigFile("file"))
}

func TestToInt_SuccessAndErrors(t *testing.T) {
	t.Parallel()

	v, err := utils.ToInt(10)
	require.NoError(t, err)
	require.Equal(t, 10, v)

	v, err = utils.ToInt(int64(20))
	require.NoError(t, err)
	require.Equal(t, 20, v)

	v, err = utils.ToInt("30")
	require.NoError(t, err)
	require.Equal(t, 30, v)

	_, err = utils.ToInt("x")
	require.ErrorIs(t, err, errors.ErrNotInt)
	_, err = utils.ToInt(float64(3.14))
	require.NoError(t, err)
}

func TestToInt32_SuccessAndErrors(t *testing.T) {
	t.Parallel()

	v, err := utils.ToInt32(int32(11))
	require.NoError(t, err)
	require.Equal(t, int32(11), v)
	v, err = utils.ToInt32(12)
	require.NoError(t, err)
	require.Equal(t, int32(12), v)
	v, err = utils.ToInt32("13")
	require.NoError(t, err)
	require.Equal(t, int32(13), v)
	_, err = utils.ToInt32("abc")
	require.ErrorIs(t, err, errors.ErrNotInt32)
	_, err = utils.ToInt32(time.Second)
	require.ErrorIs(t, err, errors.ErrNotInt32)
}

func TestToInt64_SuccessAndErrors(t *testing.T) {
	t.Parallel()

	v, err := utils.ToInt64(int64(99))
	require.NoError(t, err)
	require.Equal(t, int64(99), v)
	v, err = utils.ToInt64(77)
	require.NoError(t, err)
	require.Equal(t, int64(77), v)
	v, err = utils.ToInt64("88")
	require.NoError(t, err)
	require.Equal(t, int64(88), v)
	_, err = utils.ToInt64("bad")
	require.ErrorIs(t, err, errors.ErrNotInt64)
}

func TestToUintVariants(t *testing.T) {
	t.Parallel()

	u, err := utils.ToUint(uint(5))
	require.NoError(t, err)
	require.Equal(t, uint(5), u)
	u, err = utils.ToUint(6)
	require.NoError(t, err)
	require.Equal(t, uint(6), u)
	u, err = utils.ToUint("7")
	require.NoError(t, err)
	require.Equal(t, uint(7), u)
	_, err = utils.ToUint(-1)
	require.ErrorIs(t, err, errors.ErrNotUint)
	_, err = utils.ToUint("-2")
	require.ErrorIs(t, err, errors.ErrNotUint)

	u32, err := utils.ToUint32(uint32(3))
	require.NoError(t, err)
	require.Equal(t, uint32(3), u32)
	u32, err = utils.ToUint32(4)
	require.NoError(t, err)
	require.Equal(t, uint32(4), u32)
	u32, err = utils.ToUint32("5")
	require.NoError(t, err)
	require.Equal(t, uint32(5), u32)
	_, err = utils.ToUint32(-1)
	require.ErrorIs(t, err, errors.ErrNotUint32)
	_, err = utils.ToUint32("bad")
	require.ErrorIs(t, err, errors.ErrNotUint32)

	u64, err := utils.ToUint64(uint64(10))
	require.NoError(t, err)
	require.Equal(t, uint64(10), u64)
	u64, err = utils.ToUint64(11)
	require.NoError(t, err)
	require.Equal(t, uint64(11), u64)
	u64, err = utils.ToUint64("12")
	require.NoError(t, err)
	require.Equal(t, uint64(12), u64)
	_, err = utils.ToUint64(-1)
	require.ErrorIs(t, err, errors.ErrNotUint64)
	_, err = utils.ToUint64("bad")
	require.ErrorIs(t, err, errors.ErrNotUint64)
}

func TestFloatConverters(t *testing.T) {
	t.Parallel()

	f32, err := utils.ToFloat32(float32(1.5))
	require.NoError(t, err)
	require.InDelta(t, 1.5, f32, 0.0001)
	f32, err = utils.ToFloat32(1.25)
	require.NoError(t, err)
	require.InDelta(t, 1.25, f32, 0.0001)
	f32, err = utils.ToFloat32("2.5")
	require.NoError(t, err)
	require.InDelta(t, 2.5, f32, 0.0001)
	_, err = utils.ToFloat32("bad")
	require.ErrorIs(t, err, errors.ErrNotFloat32)

	f64, err := utils.ToFloat64(float64(3.5))
	require.NoError(t, err)
	require.InDelta(t, 3.5, f64, 0.0001)
	f64, err = utils.ToFloat64(float32(4.5))
	require.NoError(t, err)
	require.InDelta(t, 4.5, f64, 0.0001)
	f64, err = utils.ToFloat64("6.75")
	require.NoError(t, err)
	require.InDelta(t, 6.75, f64, 0.0001)
	_, err = utils.ToFloat64("bad")
	require.ErrorIs(t, err, errors.ErrNotFloat64)
}

func TestBasicConverters(t *testing.T) {
	t.Parallel()

	s, err := utils.ToString("ok")
	require.NoError(t, err)
	require.Equal(t, "ok", s)
	_, err = utils.ToString(1)
	require.ErrorIs(t, err, errors.ErrNotString)

	b, err := utils.ToBool(true)
	require.NoError(t, err)
	require.True(t, b)
	b, err = utils.ToBool("true")
	require.NoError(t, err)
	require.True(t, b)
	_, err = utils.ToBool("bad")
	require.ErrorIs(t, err, errors.ErrNotBool)

	ss, err := utils.ToStringSlice([]any{"a", "b"})
	require.NoError(t, err)
	require.Equal(t, []string{"a", "b"}, ss)
	_, err = utils.ToStringSlice([]any{"a", 2})
	require.ErrorIs(t, err, errors.ErrNotStringInSlice)
	_, err = utils.ToStringSlice(42)
	require.ErrorIs(t, err, errors.ErrNotStringSlice)

	m, err := utils.ToMap(map[string]any{"k": "v"})
	require.NoError(t, err)
	require.Equal(t, "v", m["k"])
	_, err = utils.ToMap(1)
	require.ErrorIs(t, err, errors.ErrNotMap)

	tm := time.Now()
	tt, err := utils.ToTime(tm)
	require.NoError(t, err)
	require.WithinDuration(t, tm, tt, time.Nanosecond)
	_, err = utils.ToTime(1)
	require.ErrorIs(t, err, errors.ErrNotTime)

	d := time.Second
	dur, err := utils.ToDuration(d)
	require.NoError(t, err)
	require.Equal(t, d, dur)
	_, err = utils.ToDuration(1)
	require.ErrorIs(t, err, errors.ErrNotDuration)

	bb, err := utils.ToBytes([]byte("x"))
	require.NoError(t, err)
	require.Equal(t, []byte("x"), bb)
	bb, err = utils.ToBytes("x")
	require.NoError(t, err)
	require.Equal(t, []byte("x"), bb)
	_, err = utils.ToBytes(1)
	require.ErrorIs(t, err, errors.ErrNotBytes)

	u := uuid.New()
	uid, err := utils.ToUUID(u)
	require.NoError(t, err)
	require.Equal(t, u, uid)
	uid, err = utils.ToUUID(u.String())
	require.NoError(t, err)
	require.Equal(t, u, uid)
	_, err = utils.ToUUID("bad")
	require.ErrorIs(t, err, errors.ErrNotUUID)

	parsed, err := url.Parse("https://example.com")
	require.NoError(t, err)
	uurl, err := utils.ToURL(parsed)
	require.NoError(t, err)
	require.Equal(t, parsed, uurl)
	uurl, err = utils.ToURL("https://example.com/x")
	require.NoError(t, err)
	require.Equal(t, "https://example.com/x", uurl.String())
	_, err = utils.ToURL(1)
	require.ErrorIs(t, err, errors.ErrNotURL)
}

func TestIsSupportedConfigFile_Extensions(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		file     string
		expected bool
	}{
		{"yaml", "a" + string(contract.ExtYAML), true},
		{"yml", "a" + string(contract.ExtYML), true},
		{"json", "a" + string(contract.ExtJSON), true},
		{"other", "a.toml", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expected, utils.IsSupportedConfigFile(tc.file))
		})
	}
}
