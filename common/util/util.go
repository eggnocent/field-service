package util

import (
	"os"
	"reflect"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

func BindFromJSON(dest any, filename, path string) error {
	v := viper.New()

	v.SetConfigType("json")             // Ini untuk parsing format JSON
	v.SetConfigName(stripExt(filename)) // Ambil nama tanpa .json
	v.AddConfigPath(path)

	err := v.ReadInConfig()
	if err != nil {
		return err
	}

	err = v.Unmarshal(&dest)
	if err != nil {
		logrus.Errorf("unmarshal json fail, err:%v", err)
		return err
	}

	return nil
}

func stripExt(file string) string {
	if len(file) > 5 && file[len(file)-5:] == ".json" {
		return file[:len(file)-5]
	}
	return file
}

func SetEnvFromConsulKV(v *viper.Viper) error {
	env := make(map[string]any)

	err := v.Unmarshal(&env)
	if err != nil {
		logrus.Errorf("unmarshal env fail, err:%v", err)
		return err
	}

	for k, v := range env {
		var (
			valOf = reflect.ValueOf(v)
			val   string
		)

		switch valOf.Kind() {
		case reflect.String:
			val = valOf.String()
		case reflect.Int:
			val = strconv.Itoa(int(valOf.Int()))
		case reflect.Uint:
			val = strconv.Itoa(int(valOf.Uint()))
		case reflect.Float32:
			val = strconv.Itoa(int(valOf.Float()))
		case reflect.Float64:
			val = strconv.Itoa(int(valOf.Float()))
		case reflect.Bool:
			val = strconv.FormatBool(valOf.Bool())
		}

		err = os.Setenv(k, val)
		if err != nil {
			logrus.Errorf("set env fail, err:%v", err)
			return err
		}
	}

	return nil
}

func BindFromConsul(dest any, endPoint, path string) error {
	v := viper.New()
	v.SetConfigType("json")
	err := v.AddRemoteProvider("consul", endPoint, path)
	if err != nil {
		logrus.Errorf("add remote provider fail, err:%v", err)
		return err
	}

	err = v.ReadRemoteConfig()
	if err != nil {
		logrus.Errorf("read remote config fail, err:%v", err)
		return err
	}

	err = v.Unmarshal(&dest)
	if err != nil {
		logrus.Errorf("unmarshal json fail, err:%v", err)
		return err
	}

	err = SetEnvFromConsulKV(v)
	if err != nil {
		logrus.Errorf("set env fail, err:%v", err)
		return err
	}

	return nil
}
