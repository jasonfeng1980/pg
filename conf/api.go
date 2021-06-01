package conf

// 全局的配置
var globalConf = DefaultConf
func Set(c Config){
    globalConf = c
}
func Get() Config {
    return globalConf
}

func ConfInit(root string) *YamlConf{
    return &YamlConf{
        Root: root,
    }
}