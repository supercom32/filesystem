package filesystem

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

/**
DownloadFile allows you to download a file from the internet to your local file
system.
*/
func DownloadFile(url string, filepath string) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	// Here we provide a fake 'user-agent' value so that our request looks like it's from a browser.
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Fedora; Linux x86_64; rv:52.0) Gecko/20100101 Firefox/52.0")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

/**
CopyFile allows you to copy a file from one source location to a target
destination location. In the event the operation could not be completed,
an error is returned to the user.
*/
func CopyFile(sourceFile string, destinationFile string) error {
	sourceFileStat, err := os.Stat(sourceFile)
	if err != nil {
		return err
	}
	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file.", sourceFile)
	}
	source, err := os.Open(sourceFile)
	if err != nil {
		return err
	}
	defer source.Close()
	destination, err := os.Create(destinationFile)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}

/**
WriteBytesToFile allows you to write a series of bytes to a given file.
In addition, the following information should be noted:

- In the event the file does not already exist, it will be created for you
with the permission attributes provided.

- If you pass in a permissions value of '0', the default value of 666 will
be used instead.
*/
func WriteBytesToFile(fileName string, bytesToWrite []byte, permissions int) error {
	if permissions == 0 {
		permissions = 666
	}
	perm := os.FileMode(uint32(permissions))
	err := ioutil.WriteFile(fileName, bytesToWrite, perm)
	return err
}

/**
GetFileContentsAsBytes allows you to get the contents of a file as a byte
array.
*/
func GetFileContentsAsBytes(fileName string) ([]byte, error) {
	var fileContents []byte
	var err error
	fileContents, err = ioutil.ReadFile(fileName)
	if err != nil {
		return fileContents, err
	}
	return fileContents, err
}

/**
AppendLineToFile allows you to append a line to the end of a file.
In addition, the following information should be noted:

- In the event the file does not already exist, it will be created for you
with the permission attributes provided.

- If you pass in a permissions value of '0', the default value of 666 will
be used instead.
*/
func AppendLineToFile(fileName string, lineToWrite string, permissions int) error {
	if permissions == 0 {
		permissions = 666
	}
	perm := os.FileMode(uint32(permissions))
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, perm)
	defer file.Close()
	file.WriteString(lineToWrite)
	return err
}

/**
GetListOfFiles allows you to obtain a list of files that match a given regular
expression.
*/
func GetListOfFiles(directoryPath string, regexMatcher string) ([]string, error) {
	return GetListOfDirectoryContents(directoryPath, regexMatcher,true, false)
}

/**
GetListOfDirectories allows you to obtain a list of files that match a given
regular expression.
*/
func GetListOfDirectories(directoryPath string, regexMatcher string) ([]string, error) {
	return GetListOfDirectoryContents(directoryPath, regexMatcher, false, true)
}

/**
GetListOfDirectoryContents allows you to obtain a list of files and directories
that match a given regular expression.
*/
func GetListOfDirectoryContents(directoryPath string, regexMatcher string, isFilesListed bool, isDirectoriesListed bool) ([]string, error) {
	var fileList []string
	files, err := ioutil.ReadDir(directoryPath)
	if err != nil {
		return fileList, err
	}
	for _, file := range files {
		regex := regexp.MustCompile(regexMatcher)
		match := regex.FindStringSubmatch(file.Name())
		if len(match) > 0 {
			if file.IsDir() && isDirectoriesListed {
				fileList = append(fileList, file.Name())
			}
			if !file.IsDir() && isFilesListed {
				fileList = append(fileList, file.Name())
			}
		}
	}
	return fileList, err
}

/**
GetNormalizedDirectoryPath allows you to guarantee that a directory path
is always formatted with a trailing slash at the end. This is useful when
you need to work with paths in predictable and consistent manner.
 */
func GetNormalizedDirectoryPath(directoryPath string) string {
	var normalizedDirectoryPath string = directoryPath
	if !strings.HasSuffix(directoryPath, "/")  && !strings.HasSuffix(directoryPath, "\\") {
		normalizedDirectoryPath = normalizedDirectoryPath + "/"
	}
	return normalizedDirectoryPath
}

/**
IsDirectoryEmpty allows you to detect if a directory is empty or not.
*/
func IsDirectoryEmpty(directoryName string) (bool, error) {
	file, err := os.Open(directoryName)
	if err != nil {
		return false, err
	}
	defer file.Close()
	_, err = file.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

/**
GetFileSize allows you to obtain the size of a specified file in bytes.
*/
func GetFileSize(fileName string) (int64, error){
	var fileSize int64
	file, err := os.Stat(fileName)
	if err != nil {
		return fileSize, err
	}
	fileSize = file.Size()
	return fileSize, err
}

/**
GetWorkingDirectory allows you to obtain the current working directory
where your program is executing.
*/
func GetWorkingDirectory() (string, error) {
	var parent string
	workingDirectory, err := os.Getwd()
	if err != nil {
		return parent, err
	}
	parent = filepath.Dir(workingDirectory)
	return parent, err
}

/**
GetParentDirectory allows you to obtain the container directory of
your provided path. This will be everything except the last element
of your path.
*/
func GetParentDirectory(fileOrDirectoryPath string) string {
	normalizedDirectory := fileOrDirectoryPath
	if strings.HasSuffix(fileOrDirectoryPath, "/") || strings.HasSuffix(fileOrDirectoryPath, "\\") {
		normalizedDirectory = normalizedDirectory[:len(normalizedDirectory)-1]
	}
	parentDirectory := filepath.Dir(normalizedDirectory)
	return parentDirectory
}

/**
RenameFile allows you to rename a file on your local file system. In the event
that a file with the same name already exists, it will be overwritten. Here we
explicitly do the delete so we don't depend on the 'os.Rename' behaviour of
overwriting files which may be environment dependant.
*/
func RenameFile(sourceFileName string, targetFileName string) error {
	var err error
	if sourceFileName == targetFileName {
		return err
	}
	if IsFileExists(targetFileName) {
		DeleteFile(targetFileName)
	}
	// Since Windows is case insensitive we check if the names are identical, we
	// give the file a temporary name before we request to rename it properly.
	if runtime.GOOS == "windows" && strings.ToLower(sourceFileName) == strings.ToLower(targetFileName) {
		err = os.Rename(sourceFileName, targetFileName + ".tmp")
		if err != nil {
			return err
		}
		err = os.Rename(targetFileName + ".tmp", targetFileName)
		return err
	}
	err = os.Rename(sourceFileName, targetFileName)
	return err
}

/**
IsDirectoryExists allows you to check if a directory exists or not on the file
system.
*/
func IsDirectoryExists(directoryPath string) bool {
	return isDiskEntryExists(directoryPath)
}

/**
IsFileExists allows you to check if a file exists or not on the file system.
*/
func IsFileExists(filePath string) bool {
	return isDiskEntryExists(filePath)
}

/**
isDiskEntryExists allows you to check if a valid disk entry exists on the
file system.
*/
func isDiskEntryExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

/**
DeleteFile allows you to delete a file on the file system.
*/
func DeleteFile(fileName string) error {
	err := os.Remove(fileName)
	return err
}

/**
DeleteFilesMatchingPattern allows you to delete files matching a specific
pattern. Pattern syntax is the same as the 'Match' command.
*/
func DeleteFilesMatchingPattern(fileName string ) error {
	files, err := filepath.Glob(fileName)
	if err != nil {
		return err
	}
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			return err
		}
	}
	return err
}

/**
MoveFile allows you to move a file from one location to another. This method
is an alias which simply performs a rename command, which is capable of doing
the same action.
*/
func MoveFile(sourceFile string, destinationFile string) error {
	err := RenameFile(sourceFile, destinationFile)
	return err
}

/**
CreateDirectory allows you to create a directory on your local file system.
*/
func CreateDirectory(directoryPath string, permissions uint32) error {
	perm := os.FileMode(permissions)
	err := os.MkdirAll(directoryPath, perm)
	return err
}

/**
GetFileExtension allows you to get the file extension of a file.
*/
func GetFileExtension(fileName string) string {
	extension := filepath.Ext(fileName)
	return extension
}

/**
GetBaseFileName allows you to extract the base name of a file without any path or
file extensions.
*/
func GetBaseFileName(fileName string) string {
	normalizedFileName := GetFileNameFromPath(fileName)
	return strings.TrimSuffix(normalizedFileName, filepath.Ext(normalizedFileName))
}

/**
GetBaseDirectory allows you to extract the directory from a file path.
*/
func GetBaseDirectory(filePath string) string {
	directory, _ := filepath.Split(filePath)
	return directory
}

/**
GetFileNameFromPath allows you to extract a file name from a fully qualified file path.
*/
func GetFileNameFromPath(fullyQualifiedFileName string) string {
	return filepath.Base(fullyQualifiedFileName)
}

/**
IsFile allows you to check if an item on the file system is a file or not.
*/
func IsFile(path string) (bool, error){
	var isFile bool
	fi, err := os.Stat(path)
	if err != nil {
		return isFile, err
	}
	mode := fi.Mode()
	if mode.IsRegular() {
		isFile = true
	}
	return isFile, err
}