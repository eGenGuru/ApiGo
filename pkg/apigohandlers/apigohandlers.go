package apigohandlers

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/elaugier/ApiGo/pkg/apigohelpers"
	"github.com/elaugier/ApiGo/pkg/apigolib"
	"github.com/spf13/viper"

	"github.com/gin-gonic/gin"
)

//Config ...
var Config viper.Viper

//RoutesConfigs ...
var RoutesConfigs map[int]*viper.Viper

//SynchronousJob ...
func SynchronousJob(c *gin.Context) {
	apigolib.Trace()
	currentRoute, j, t, err := params(c)
	if err != nil {
		return
	}
	k := apigohelpers.NewKafka()
	err = k.Send(j, t)
	if err != nil {
		errorMsg := fmt.Sprintf("Error on send synchonous job : %v", err)
		log.Println(errorMsg)
		c.JSON(500, gin.H{
			"msg": fmt.Sprintln(errorMsg),
		})
		return
	}
	c.JSON(200, gin.H{
		"msg": fmt.Sprintf("%s", currentRoute),
	})
}

//AsynchronousJob ...
func AsynchronousJob(c *gin.Context) {
	apigolib.Trace()
	currentRoute, j, t, err := params(c)
	if err != nil {
		return
	}
	k := apigohelpers.NewKafka()
	err = k.Send(j, t)
	if err != nil {
		errorMsg := fmt.Sprintf("Error on send synchonous job : %v", err)
		log.Println(errorMsg)
		c.JSON(500, gin.H{
			"msg": fmt.Sprintln(errorMsg),
		})
		return
	}
	c.JSON(200, gin.H{
		"msg": fmt.Sprintf("%s", currentRoute),
	})
}

func params(c *gin.Context) (string, apigohelpers.JSONCmd, string, error) {
	buf, _ := c.Get("id")
	id := buf.(int)
	var Route RouteConfig
	err := RoutesConfigs[id].Unmarshal(&Route)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}
	currentRoute := Route.Name
	log.Printf("Current Route => %s", currentRoute)
	currentParams := make(map[string]string)
	var value string
	for i := 0; i < len(Route.Cmd.Params); i++ {

		p := Route.Cmd.Params[i]

		log.Printf("Expected parameter name: %s", p.Name)
		mandatory, err := strconv.ParseBool(p.Mandatory)
		if err == nil {
			switch p.In {
			case "uri":
				value = c.Param(p.Name)
				log.Printf("retrieve key '%s' => '%s' from %s", p.Name, value, p.In)

				if value == "" && mandatory {
					errMsg := fmt.Sprintf("Parameter '%s' is mandatory => raise error and add message for response", p.Name)
					log.Printf(errMsg)
					c.JSON(400, gin.H{
						"msg": errMsg,
					})
					return "", apigohelpers.JSONCmd{}, Route.Topic, errors.New(errMsg)
				}

				if value == "" && !mandatory {
					log.Printf("Parameter '%s' is not mandatory but empty => no action", p.Name)
				}

				if !IsValueTypeOfExpected(value, p.Type) {
					errMsg := fmt.Sprintf("Parameter '%s' : bad type", p.Name)
					c.JSON(400, gin.H{
						"msg": errMsg,
					})
					return "", apigohelpers.JSONCmd{}, Route.Topic, errors.New(errMsg)
				}

			case "header":
				value = c.GetHeader(p.Name)
				log.Printf("retrieve key '%s' => '%s' from %s", p.Name, value, p.In)

				if value == "" && mandatory {
					errMsg := fmt.Sprintf("Parameter '%s' is mandatory => raise error and add message for response", p.Name)
					log.Printf(errMsg)
					c.JSON(400, gin.H{
						"msg": errMsg,
					})
					return "", apigohelpers.JSONCmd{}, Route.Topic, errors.New(errMsg)
				}

				if value == "" && !mandatory {
					log.Printf("Parameter '%s' is not mandatory but empty => no action", p.Name)
				}

				if !IsValueTypeOfExpected(value, p.Type) {
					errMsg := fmt.Sprintf("Parameter '%s' : bad type", p.Name)
					c.JSON(400, gin.H{
						"msg": errMsg,
					})
					return "", apigohelpers.JSONCmd{}, Route.Topic, errors.New(errMsg)
				}

			case "querystring":
				value = c.Query(p.Name)
				log.Printf("retrieve key '%s' => '%s' from %s", p.Name, value, p.In)

				if value == "" && mandatory {
					errMsg := fmt.Sprintf("Parameter '%s' is mandatory => raise error and add message for response", p.Name)
					log.Printf(errMsg)
					c.JSON(400, gin.H{
						"msg": errMsg,
					})
					return "", apigohelpers.JSONCmd{}, Route.Topic, errors.New(errMsg)
				}

				if value == "" && !mandatory {
					log.Printf("Parameter '%s' is not mandatory but empty => no action", p.Name)
				}

				if !IsValueTypeOfExpected(value, p.Type) {
					errMsg := fmt.Sprintf("Parameter '%s' : bad type", p.Name)
					c.JSON(400, gin.H{
						"msg": errMsg,
					})
					return "", apigohelpers.JSONCmd{}, Route.Topic, errors.New(errMsg)
				}

			case "body":
				var keyValue map[string]string
				c.BindJSON(&keyValue)
				value = keyValue[p.Name]
				log.Printf("retrieve key '%s' => '%s' from %s", p.Name, value, p.In)

				if value == "" && mandatory {
					errMsg := fmt.Sprintf("Parameter '%s' is mandatory => raise error and add message for response", p.Name)
					log.Printf(errMsg)
					c.JSON(400, gin.H{
						"msg": errMsg,
					})
					return "", apigohelpers.JSONCmd{}, Route.Topic, errors.New(errMsg)
				}

				if value == "" && !mandatory {
					log.Printf("Parameter '%s' is not mandatory but empty => no action", p.Name)
				}

				if !IsValueTypeOfExpected(value, p.Type) {
					errMsg := fmt.Sprintf("Parameter '%s' : bad type", p.Name)
					c.JSON(400, gin.H{
						"msg": errMsg,
					})
					return "", apigohelpers.JSONCmd{}, Route.Topic, errors.New(errMsg)
				}

			default:
				errMsg := fmt.Sprintf("Unkown 'In' value for param '%s'", p.Name)
				log.Printf(errMsg)
				c.JSON(500, gin.H{
					"msg": errMsg,
				})
				return "", apigohelpers.JSONCmd{}, Route.Topic, errors.New(errMsg)
			}
		} else {
			errMsg := fmt.Sprintf("Error while parsing Mandatory option for param %s", p.Name)
			log.Printf(errMsg)
			c.JSON(500, gin.H{
				"msg": errMsg,
			})
			return "", apigohelpers.JSONCmd{}, Route.Topic, errors.New(errMsg)
		}
		currentParams[p.Name] = value
	}
	jid, err := uuid.NewUUID()
	if err != nil {
		log.Printf("Error on generating jid : %v", err)
		return "", apigohelpers.JSONCmd{}, Route.Topic, err
	}
	mid := jid.String()
	j := apigohelpers.JSONCmd{
		UUID:     mid,
		Name:     Route.Cmd.Name,
		Type:     Route.Cmd.Type,
		PSModule: Route.Cmd.PSModule,
		PyVenv:   Route.Cmd.PyVenv,
		Params:   currentParams,
		JobType:  Route.JobType,
		Timeout:  Route.Timeout,
	}
	return "", j, Route.Topic, nil
}

//GetJobStatus ...
func GetJobStatus(c *gin.Context) {
	UUID := c.Param("uuid")
	c.JSON(200, gin.H{
		"msg": fmt.Sprintf("%s", UUID),
	})
}

//Ping ...
func Ping(version string) gin.HandlerFunc {
	apigolib.Trace()
	return func(c *gin.Context) {
		c.JSON(200, gin.H{
			"msg": fmt.Sprintf("pong %s", version),
		})
	}
}

//PageNotFound ...
func PageNotFound(c *gin.Context) {
	c.JSON(404, gin.H{
		"sts": "failed",
		"hco": "404",
		"msg": "PageNotFound",
	})
}

//MethodNotAllowed ...
func MethodNotAllowed(c *gin.Context) {
	c.JSON(405, gin.H{
		"sts": "failed",
		"hco": "405",
		"msg": "MethodNotAllowed",
	})
}

//IsValueTypeOfExpected ...
func IsValueTypeOfExpected(value string, typeExpected string) bool {
	switch strings.ToLower(typeExpected) {
	case "int":
		_, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return false
		}
	case "float":
		_, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return false
		}
	case "uint":
		_, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return false
		}
	case "string":
	default:
		return true
	}
	return true
}

//Parameter ...
type Parameter struct {
	Name      string `json:"Name"`
	Type      string `json:"Type"`
	Mandatory string `json:"Mandatory"`
	In        string `json:"In"`
}

//Cmd ...
type Cmd struct {
	Name     string      `json:"Name"`
	Type     string      `json:"Type"`
	PSModule string      `json:"PSModule"`
	PyVenv   string      `json:"PyVenv"`
	Params   []Parameter `json:"Params"`
}

//RouteConfig ...
type RouteConfig struct {
	Name              string `json:"Name"`
	Cmd               Cmd    `json:"Cmd"`
	Route             string `json:"Route"`
	Method            string `json:"Method"`
	JobType           string `json:"JobType"`
	Topic             string `json:"Topic"`
	Timeout           string `json:"Timeout"`
	AddRequestIDParam string `json:"AddRequestIDParam"`
}
