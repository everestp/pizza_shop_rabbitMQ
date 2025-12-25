package config

import (
    "fmt"
    "os"
    "reflect"

    "github.com/everestp/pizza-shop/logger"
    "github.com/joho/godotenv"
)

// 1. The Global State
// 'env' holds all your configuration in memory so you don't have to 
// read from the disk/system every time you need a variable.
var env ConfigDto

// 2. The Blueprint (DTO)
// Using a struct ensures that you have a "list" of expected variables.
// Note: Fields start with lowercase, meaning they are private to this package.
type ConfigDto struct {
    port                    string
    rabbit_mq_host          string
    rabbit_mq_username      string
    rabbit_mq_password      string
    rabbit_mq_port          string
    rabbit_mq_default_queue string
}

// 3. The Loader
// This function maps the raw strings from os.Getenv into our structured struct.
func ConfigEnv() {
    LoadEnvVariable()
    env = ConfigDto{
        port:                    os.Getenv("PORT"),
        rabbit_mq_host:          os.Getenv("RABBIT_MQ_HOST"),
        rabbit_mq_username:      os.Getenv("RABBIT_MQ_USERNAME"),
        rabbit_mq_password:      os.Getenv("RABBIT_MQ_PASSWORD"),
        rabbit_mq_port:          os.Getenv("RABBIT_MQ_PORT"),
        rabbit_mq_default_queue: os.Getenv("RABBIT_MQ_DEFAULT_QUEUE"),
    }
}

// 4. The Auto-Runner (init)
// Go runs 'init' functions automatically when the package is imported.
// This ensures that 'env' is never empty when the app starts.
func init() {
    if env.port == "" {
        ConfigEnv()
    }
}

// 5. Reflection Logic (The "Magic" Finder)
// This uses 'reflect' to look up a struct field by its string name.
// Example: accessField("port") will look for env.port.
func accessField(key string) (string, error) {
    v := reflect.ValueOf(env)
    t := v.Type()

    if t.Kind() != reflect.Struct {
        return "", fmt.Errorf("expected struct")
    }

    // Check if the field actually exists in the ConfigDto struct
    _, ok := t.FieldByName(key)
    if !ok {
        return "", fmt.Errorf("field not found: %s", key)
    }

    // Return the string value of that field
    f := v.FieldByName(key)
    return f.String(), nil
}

// 6. Dotenv Loader
// This looks for a file named '.env'. If it finds it, it loads it.
// If it doesn't (like in production/Docker), it ignores it and uses system envs.
func LoadEnvVariable() {
    if _, err := os.Stat(".env"); err == nil {
        err = godotenv.Load()
        if err != nil {
            panic("error loading .env file")
        }
    } else {
        logger.Log("No .env file found; using system environment variables")
    }
}

// 7. The Public API
// This is the function you call from other packages.
// Usage: config.GetEnvProperty("rabbit_mq_host")
func GetEnvProperty(propertyKey string) string {
    // Safety check: if env is empty, load it again
    if env.port == "" {
        ConfigEnv()
    }
    
    val, err := accessField(propertyKey)
    if err != nil {
        logger.Log(fmt.Sprintf("Error accessing config field: %v", propertyKey))
    }
    return val
}