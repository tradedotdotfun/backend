package utils

import (
	"encoding/json"
	"log"
	"reflect"

	"github.com/gofiber/fiber/v2"
)

func Wrap(fn func(*fiber.Ctx) (interface{}, error)) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		result, err := fn(c)
		if err != nil {
			return err
		}

		v := reflect.ValueOf(result)

		if v.Kind() == reflect.Ptr {
			v = reflect.Indirect(v)
		}

		if v.Kind() == reflect.Map || v.Kind() == reflect.Struct {
			body, err := json.Marshal(result)
			if err != nil {
				log.Printf("Fail to Encode Json(%+v) %+v\n", result, err)
				return fiber.ErrInternalServerError
			}
			c.Send(body)
		} else if v.Kind() == reflect.Slice {
			if v.Type() == reflect.TypeOf([]byte(nil)) {
				c.Send(result.([]byte))
			} else {
				body, err := json.Marshal(result)
				if err != nil {
					log.Printf("Fail to Encode Json(%+v) %+v\n", result, err)
					return fiber.ErrInternalServerError
				}
				c.Send(body)
			}
		}

		return nil
	}
}
