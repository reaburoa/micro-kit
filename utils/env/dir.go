package env

import (
	"os"
	"path/filepath"
)

// 单测时候需要一些本地文件路径的处理，只用在单测中, 本地启动可能也会调用
// 项目名字，需要在项目启动的时候设置
// 取的顺序，
// 如果设置环境变量PROJECT_ROOT_PATH，就用环境变量的
// 如果没有设置环境变量，就从运行路径中获取这个listenbook-appability后面的第一个目录
// 如果都没有，就用no_set_project_name
var rootPath = ""
var rootPathEnv = "PROJECT_ROOT_PATH"

func GetProjectPath() (string, error) {
	if rootPath != "" {
		return rootPath, nil
	}
	// 先从环境变量中获取
	p := os.Getenv(rootPathEnv)
	if p != "" {
		rootPath = p
		return rootPath, nil
	}
	// 从运行路径获取
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// 遍历所有路径找到configs目录，configs目录所在的地方就是项目根目录
	rootPath = findDir(pwd, "configs")
	return rootPath, nil
}

func findDir(checkPath, checkDir string) string {
	for {
		// 检查当前目录下是否存在 configs 子目录
		configPath := filepath.Join(checkPath, checkDir)
		if _, err := os.Stat(configPath); err == nil {
			// 找到 configs 目录，返回该目录的路径
			return checkPath
		} else if !os.IsNotExist(err) {
			// 发生了除了文件不存在以外的错误
			return ""
		}

		// 未找到 configs 目录，向上移动到父目录
		parentDir := filepath.Dir(checkPath)
		if parentDir == checkPath {
			// 如果父目录与当前目录相同，说明已经到达根目录，停止搜索
			break
		}
		checkPath = parentDir
	}
	return ""
}

// RestRootPath reset root path
// 一些特殊场景可能需要,提供一个口子
func RestRootPath(restPath string) {
	rootPath = restPath
}
