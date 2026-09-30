package filesystem

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
)

type fileInstanceType struct {
	fileDescriptor *os.File
}

/*
GetFileInstance allows you to obtain a file instance to work on.
*/
func GetFileInstance() fileInstanceType {
	var fileInstance fileInstanceType
	return fileInstance
}

/*
Open allows you to access a file on the file system in the open state.
*/
func (shared *fileInstanceType) Open(fileName string, permissions int) error {
	if permissions == 0 {
		permissions = 0644
	}
	perm := os.FileMode(uint32(permissions))
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, perm)
	if err != nil {
		return err
	}
	shared.fileDescriptor = file
	return err
}

/*
Close allows you to close a file which has already been opened.
*/
func (shared *fileInstanceType) Close() {
	if shared.fileDescriptor == nil {
		panic("There is no open file to close.")
	}
	shared.fileDescriptor.Close()
}

/*
WriteBytes allows you to add an arbitrary number of bytes to an open file.
*/
func (shared *fileInstanceType) WriteBytes(bytes []byte) error {
	if shared.fileDescriptor == nil {
		panic("There is no open file for writing bytes to.")
	}
	_, err := shared.fileDescriptor.Write(bytes)
	return err
}

/*
WriteLine allows you to add string data to an open file as a line. A newline
identifier will automatically be added to your string.
*/
func (shared *fileInstanceType) WriteLine(lineToWrite string) error {
	if shared.fileDescriptor == nil {
		panic("There is no open file for writing lines to.")
	}
	err := shared.WriteString(lineToWrite + "\n")
	return err
}

/*
WriteString allows you to add string data to an open file.
*/
func (shared *fileInstanceType) WriteString(stringToWrite string) error {
	if shared.fileDescriptor == nil {
		panic("There is no open file for writing strings to.")
	}
	_, err := shared.fileDescriptor.WriteString(stringToWrite)
	return err
}

/*
GetFileContents allows you to get the entire file contents.
*/
func (shared *fileInstanceType) GetFileContents() ([]byte, error) {
	if shared.fileDescriptor == nil {
		panic("There is no open file for reading with.")
	}
	fileInfo, err := shared.fileDescriptor.Stat()
	if err != nil {
		return nil, err
	}
	buffer := make([]byte, fileInfo.Size())
	_, err = shared.fileDescriptor.ReadAt(buffer, 0)
	if err != nil {
		return nil, err
	}
	formattedBuffer := bytes.TrimRight(buffer, "\n")
	return formattedBuffer, err
}

/*
GetFirstLine allows you to get the first line from a text file.
*/
func (shared *fileInstanceType) GetFirstLine() ([]byte, error) {
	fileInfo, err := shared.fileDescriptor.Stat()
	if err != nil {
		return nil, err
	}
	_, err = shared.fileDescriptor.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}
	buffer := bytes.NewBuffer(make([]byte, 0, fileInfo.Size()))
	_, err = io.Copy(buffer, shared.fileDescriptor)
	if err != nil {
		return nil, err
	}
	firstLine, err := buffer.ReadBytes('\n')
	if err != nil && err != io.EOF {
		return nil, err
	}
	_, err = shared.fileDescriptor.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}
	// Remove trailing delimiter "\n" or "\r\n" in a UTF-8 safe way
	firstLine = bytes.TrimSuffix(firstLine, []byte{'\n'})
	firstLine = bytes.TrimSuffix(firstLine, []byte{'\r'})
	return firstLine, err
}

/*
RemoveFirstLine allows you to remove the first line from a text file.
This function properly handles UTF-8 encoded text and different line endings.
*/
func (shared *fileInstanceType) RemoveFirstLine() error {
	fileInfo, err := shared.fileDescriptor.Stat()
	if err != nil {
		return err
	}
	_, err = shared.fileDescriptor.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	// Read the entire file content
	fileContent := make([]byte, fileInfo.Size())
	_, err = shared.fileDescriptor.Read(fileContent)
	if err != nil {
		return err
	}

	// Create a scanner to properly handle UTF-8 and different line endings
	scanner := bufio.NewScanner(bytes.NewReader(fileContent))

	// Skip the first line
	if !scanner.Scan() {
		// If there's no first line (empty file), just return
		if scanner.Err() != nil {
			return scanner.Err()
		}
		return nil
	}

	// Get the remaining content
	var remainingContent bytes.Buffer
	for scanner.Scan() {
		remainingContent.WriteString(scanner.Text() + "\n")
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// Truncate the file and write the remaining content
	err = shared.fileDescriptor.Truncate(0)
	if err != nil {
		return err
	}
	_, err = shared.fileDescriptor.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	_, err = shared.fileDescriptor.Write(remainingContent.Bytes())
	if err != nil {
		return err
	}
	err = shared.fileDescriptor.Sync()
	if err != nil {
		return err
	}
	_, err = shared.fileDescriptor.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	return nil
}

/*
GetDefaultCacheDirectory allows you to obtain the cache directory of a user.
The cache directory is useful for storing program data that needs to
be accessed frequently, but is not terribly important in case it has to be
regenerated.
*/
func GetDefaultCacheDirectory() (string, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not get user home directory: %v", err)
	}
	switch runtime.GOOS {
	case "linux":
		return filepath.Join(userHomeDir, ".cache"), nil
	case "windows":
		return filepath.Join(userHomeDir, "AppData", "Local"), nil
	case "darwin":
		return filepath.Join(userHomeDir, "Library", "Caches"), nil
	}
	return "", errors.New("could not determine cache directory")
}

/*
IsFileContainsText allows you to check if a file contains a matching string or not. Since this method loads the entire
contents of a file into memory, it should only be used for smaller files.
*/
func IsFileContainsText(filename string, regexMatcher string) (bool, error) {
	fileContents, err := ioutil.ReadFile(filename)
	if err != nil {
		return false, err
	}
	regex := regexp.MustCompile(regexMatcher)
	match := regex.Find(fileContents)
	if len(match) > 0 {
		return true, err
	}
	return false, err
}

/*
FindReplaceInFile allows you to find and replace text from within a file. Since this method loads the entire
contents of a file into memory, it should only be used for smaller files.
*/
func FindReplaceInFile(filename string, regexMatcher string, replacementValue string) error {
	fileContents, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	fileContentString := string(fileContents)
	lines := strings.Split(fileContentString, "\n")
	regex := regexp.MustCompile(regexMatcher)
	for i, line := range lines {
		lines[i] = regex.ReplaceAllString(line, replacementValue)
	}
	newContents := []byte(strings.Join(lines, "\n"))
	err = ioutil.WriteFile(filename, newContents, 0)
	if err != nil {
		return err
	}
	return nil
}

/*
GetLastLineFromFile allows you to obtain the last line from a file.
This function properly handles UTF-8 encoded text and different line endings.
*/
func GetLastLineFromFile(fileName string) (string, error) {
	fileContents, err := GetFileContents(fileName)
	if err != nil {
		return "", err
	}

	// Handle empty files
	if len(fileContents) == 0 {
		return "", nil
	}

	// Use bufio.Scanner which properly handles UTF-8 and different line endings
	scanner := bufio.NewScanner(bytes.NewReader(fileContents))
	var lastLine string

	for scanner.Scan() {
		lastLine = scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return lastLine, nil
}

/*
GetFileContents allows you to get the entire contents of a file.
*/
func GetFileContents(fileName string) ([]byte, error) {
	file := GetFileInstance()
	err := file.Open(fileName, 0)
	if err != nil {
		return nil, err
	}
	fileContents, err := file.GetFileContents()
	if err != nil {
		return nil, err
	}
	file.Close()
	return fileContents, err
}

/*
RemoveFirstLineFromFile allows you to remove the first line from a file.
*/
func RemoveFirstLineFromFile(fileName string) error {
	file := GetFileInstance()
	err := file.Open(fileName, 0)
	if err != nil {
		return err
	}
	err = file.RemoveFirstLine()
	if err != nil {
		return err
	}
	file.Close()
	return err
}

/*
*
DownloadFile allows you to download a file from the internet to your local.com file
system.
*/
func DownloadFile(url string, filepath string, header http.Header) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	if header == nil {
		// Here we provide a fake 'user-agent' value so that our request looks like it's from a browser.
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Fedora; Linux x86_64; rv:52.0) Gecko/20100101 Firefox/52.0")
	} else {
		req.Header = header
	}
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

/*
*
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
	defer source.Close()
	if err != nil {
		return err
	}

	destination, err := os.Create(destinationFile)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}

/*
*
WriteBytesToFile allows you to write a series of bytes to a given file.
In addition, the following information should be noted:

- In the event the file does not already exist, it will be created for you
with the permission attributes provided.

- If you pass in a permissions value of '0', the default value of 666 will
be used instead.
*/
func WriteBytesToFile(fileName string, bytesToWrite []byte, permissions int) error {
	if permissions == 0 {
		permissions = 0744
	}
	perm := os.FileMode(permissions)
	err := ioutil.WriteFile(fileName, bytesToWrite, perm)
	return err
}

/*
*
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

/*
*
AppendLineToFile allows you to append a line to the end of a file.
In addition, the following information should be noted:

- In the event the file does not already exist, it will be created for you
with the permission attributes provided.

- If you pass in a permissions value of '0', the default value of 666 will
be used instead.
*/
func AppendLineToFile(fileName string, lineToWrite string, permissions int) error {
	if permissions == 0 {
		permissions = 0644
	}
	perm := os.FileMode(uint32(permissions))
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, perm)
	defer file.Close()
	file.WriteString(lineToWrite)
	return err
}

/*
*
GetListOfFiles allows you to obtain a list of files that match a given regular
expression.
*/
func GetListOfFiles(directoryPath string, regexMatcher string) ([]string, error) {
	return GetListOfDirectoryContents(directoryPath, []string{regexMatcher}, true, false)
}

/*
*
GetListOfDirectories allows you to obtain a list of files that match a given
regular expression.
*/
func GetListOfDirectories(directoryPath string, regexMatcher string) ([]string, error) {
	return GetListOfDirectoryContents(directoryPath, []string{regexMatcher}, false, true)
}

/*
IsDirectory allows you to check if a disk entry is a directory or not.
*/
func IsDirectory(directoryPath string) bool {
	file, err := os.Open(directoryPath)
	defer file.Close()
	if err != nil {
		return false
	}
	fileInfo, err := file.Stat()
	if fileInfo.IsDir() {
		return true
	}
	return false
}

/*
*
GetListOfDirectoryContents allows you to obtain a list of files and directories
that match a given regular expression.
*/
func GetListOfDirectoryContents(directoryPath string, regexMatchers []string, isFilesIncluded bool, isDirectoriesIncluded bool) ([]string, error) {
	var fileList []string
	bareDirectoryPath := GetBareDirectoryPath(directoryPath)
	files, err := ioutil.ReadDir(bareDirectoryPath)
	if err != nil {
		return fileList, err
	}
	for _, file := range files {
		for _, currentRegex := range regexMatchers {
			regex := regexp.MustCompile(currentRegex)
			match := regex.FindStringSubmatch(file.Name())
			if len(match) > 0 {
				if file.IsDir() && isDirectoriesIncluded {
					fileList = append(fileList, file.Name()+"/")
					break
				}
				if !file.IsDir() && isFilesIncluded {
					fileList = append(fileList, file.Name())
					break
				}
			}
		}
	}
	return fileList, err
}

/*
*
FindMatchingContent allows you to find matching content from a given directory
path. Both shallow and recursive searches are supported and results are
returned as a fully qualified path.
*/
func FindMatchingContent(directoryPath string, regexMatchers []string, isFilesIncluded bool, isDirectoriesIncluded bool, isRecursive bool) ([]string, error) {
	var err error
	var listOfContents []string
	if isRecursive {
		err = filepath.Walk(directoryPath,
			func(path string, fileInfo os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !IsDirectory(path) {
					return nil
				}
				normalizedPath := GetNormalizedDirectoryPath(path)
				matchingContents, err := GetListOfDirectoryContents(normalizedPath, regexMatchers, isFilesIncluded, isDirectoriesIncluded)
				if err != nil {
					return err
				}
				matchingContents = addPrefixToStrings(normalizedPath, matchingContents)
				listOfContents = append(listOfContents, matchingContents...)
				return nil
			})
	} else {
		matchingContent, err := GetListOfDirectoryContents(directoryPath, regexMatchers, isFilesIncluded, isDirectoriesIncluded)
		if err != nil {
			return listOfContents, err
		}
		normalizedPath := GetNormalizedDirectoryPath(directoryPath)
		listOfContents = addPrefixToStrings(normalizedPath, matchingContent)
	}
	return listOfContents, err
}

/*
*
addPrefixToStrings allows you to append a prefix to an array of strings.
*/
func addPrefixToStrings(prefixToAdd string, stringArray []string) []string {
	for currentStringIndex := 0; currentStringIndex < len(stringArray); currentStringIndex++ {
		stringArray[currentStringIndex] = prefixToAdd + stringArray[currentStringIndex]
	}
	return stringArray
}

/*
*
GetNormalizedDirectoryPath allows you to guarantee that a directory path
is always formatted with a trailing slash at the end. This is useful when
you need to work with paths in predictable and consistent manner.
*/
func GetNormalizedDirectoryPath(directoryPath string) string {
	normalizedDirectoryPath := directoryPath
	if !strings.HasSuffix(directoryPath, "/") && !strings.HasSuffix(directoryPath, "\\") {
		normalizedDirectoryPath = normalizedDirectoryPath + "/"
	}
	return normalizedDirectoryPath
}

/*
GetBareDirectoryPath allows you to get a directory path without a trailing
slash. This is useful for functions which explicitly need directories
formatted this way.
*/
func GetBareDirectoryPath(directoryPath string) string {
	bareDirectoryPath := directoryPath
	if strings.HasSuffix(directoryPath, "/") || strings.HasSuffix(directoryPath, "\\") {
		bareDirectoryPath = strings.TrimSuffix(directoryPath, "/")
		bareDirectoryPath = strings.TrimSuffix(bareDirectoryPath, "\\")
	}
	return bareDirectoryPath
}

/*
*
IsDirectoryEmpty allows you to detect if a directory is empty or not.
*/
func IsDirectoryEmpty(directoryName string) (bool, error) {
	file, err := os.Open(directoryName)
	defer file.Close()
	if err != nil {
		return false, err
	}
	_, err = file.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

/*
*
GetFileSize allows you to obtain the size of a specified file in bytes.
*/
func GetFileSize(fileName string) (int64, error) {
	var fileSize int64
	file, err := os.Stat(fileName)
	if err != nil {
		return fileSize, err
	}
	fileSize = file.Size()
	return fileSize, err
}

/*
*
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

/*
*
GetParentDirectory allows you to obtain the container directory of
your provided path. This will be everything except the last element
of your path.
*/
func GetParentDirectory(fileOrDirectoryPath string) string {
	normalizedDirectory := fileOrDirectoryPath
	if strings.HasSuffix(fileOrDirectoryPath, "/") {
		normalizedDirectory = strings.TrimSuffix(normalizedDirectory, "/")
	} else if strings.HasSuffix(fileOrDirectoryPath, "\\") {
		normalizedDirectory = strings.TrimSuffix(normalizedDirectory, "\\")
	}
	parentDirectory := filepath.Dir(normalizedDirectory)
	return parentDirectory
}

/*
*
RenameFile allows you to rename a file on your local.com file system. In the event
that a file with the same name already exists, it will be overwritten. Here we
explicitly do the delete so we don't depend on the 'os.Rename' behaviour of
overwriting files which may be environment dependant.
*/
/*func RenameFile(sourceFileName string, targetFileName string) error {
	var err error
	bareSourceFileName := GetBareDirectoryPath(sourceFileName)
	bareTargetFileName := GetBareDirectoryPath(targetFileName)
	if bareSourceFileName == bareTargetFileName {
		return err
	}
	// Since Windows is case-insensitive we check if the names are identical, we
	// give the file a temporary name before we request to rename it properly.
	if runtime.GOOS == "windows" && strings.ToLower(bareSourceFileName) == strings.ToLower(bareTargetFileName) {
		err = os.Rename(bareSourceFileName, bareTargetFileName+".tmp")
		if err != nil {
			return err
		}
		err = os.Rename(bareTargetFileName+".tmp", bareTargetFileName)
		return err
	}
	// This needs to be here to avoid deleting target files on Windows before a case-insensitive check is done.
	if IsFileExists(bareTargetFileName) {
		DeleteFile(bareTargetFileName)
	}
	err = os.Rename(bareSourceFileName, bareTargetFileName)
	return err
}
*/

/*
*
RenameFile allows you to rename a file on your local.com file system. In the event
that a file with the same name already exists, it will be overwritten. Here we
explicitly do the delete so we don't depend on the 'os.Rename' behaviour of
overwriting files which may be environment dependant.
*/
func RenameFile(sourceFileName string, targetFileName string) error {
	var err error
	bareSourceFileName := GetBareDirectoryPath(sourceFileName)
	bareTargetFileName := GetBareDirectoryPath(targetFileName)

	if bareSourceFileName == bareTargetFileName {
		return nil
	}

	if runtime.GOOS == "windows" && strings.ToLower(bareSourceFileName) == strings.ToLower(bareTargetFileName) {
		err = os.Rename(bareSourceFileName, bareTargetFileName+".tmp")
		if err != nil {
			return err
		}
		err = os.Rename(bareTargetFileName+".tmp", bareTargetFileName)
		return err
	}

	var backupFileName string
	if IsFileExists(bareTargetFileName) {
		backupFileName = bareTargetFileName + ".backup"
		err = os.Rename(bareTargetFileName, backupFileName)
		if err != nil {
			return fmt.Errorf("failed to rename existing target file %s to backup %s: %w", bareTargetFileName, backupFileName, err)
		}
	}

	err = os.Rename(bareSourceFileName, bareTargetFileName)
	if err != nil && backupFileName != "" {
		rollbackErr := os.Rename(backupFileName, bareTargetFileName)
		if rollbackErr != nil {
			return fmt.Errorf("failed to restore backup file %s: %w", backupFileName, rollbackErr)
		}
		return fmt.Errorf("failed to rename %s to %s: %w", bareSourceFileName, bareTargetFileName, err)
	}

	if backupFileName != "" {
		err := DeleteFile(backupFileName)
		if err != nil {
			return fmt.Errorf("rename succeeded but failed to delete backup file %s: %w", backupFileName, err)
		}
	}

	return nil
}

/*
*
IsDirectoryExists allows you to check if a directory exists or not on the file
system.
*/
func IsDirectoryExists(directoryPath string) bool {
	return isDiskEntryExists(directoryPath)
}

/*
*
IsFileExists allows you to check if a file exists or not on the file system.
*/
func IsFileExists(filePath string) bool {
	return isDiskEntryExists(filePath)
}

/*
*
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

func GetAbsolutePath(filePath string) (string, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}
	return absPath, nil
}

/*
*
DeleteFile allows you to delete a file on the file system.
*/
func DeleteFile(fileName string) error {
	err := os.Remove(fileName)
	return err
}

/*
*
DeleteFilesMatchingPattern allows you to delete files matching a specific
pattern. Pattern syntax is the same as the 'Match' command.
*/
func DeleteFilesMatchingPattern(fileName string) error {
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

/*
DeleteDirectory allows you to recursively remove a directory from the file
system.
*/
func DeleteDirectory(pathName string) error {
	err := os.RemoveAll(pathName)
	return err
}

/*
*
MoveFile allows you to move a file from one location to another. This method
is an alias which simply performs a rename command, which is capable of doing
the same action.
*/
func MoveFile(sourceFile string, destinationFile string) error {
	// Since windows is case-insensitive, it is valid if the user wants to rename / move a file to the
	// same location, since he is just trying to change case. So we only error out if this edge-case is not detected.
	if runtime.GOOS != "windows" && sourceFile != destinationFile && IsFileExists(destinationFile) {
		return errors.New(fmt.Sprintf("Cannot move '%s' to the destination location since '%s' already exists .", sourceFile, destinationFile))
	}
	if runtime.GOOS == "windows" && strings.ToLower(sourceFile) != strings.ToLower(destinationFile) && IsFileExists(destinationFile) {
		return errors.New(fmt.Sprintf("Cannot move '%s' to the destination location since '%s' already exists .", sourceFile, destinationFile))
	}
	now := time.Now()
	uniqueId := fmt.Sprintf("%d", now.Unix())
	uniqueFileName := destinationFile + "_" + uniqueId
	err := RenameFile(sourceFile, uniqueFileName)
	if err != nil {
		return err
	}
	err = RenameFile(uniqueFileName, destinationFile)
	return err
}

func MoveDirectories(sourceDir string, destinationDir string) error {
	normalizedSourceDirectory := strings.TrimRight(sourceDir, "\\/")
	normalizedDestinationDirectory := strings.TrimRight(destinationDir, "\\/")
	if runtime.GOOS != "windows" && sourceDir != destinationDir && IsDirectoryExists(destinationDir) {
		return errors.New(fmt.Sprintf("Cannot move '%s' to the destination location since '%s' already exists.", sourceDir, destinationDir))
	}
	if runtime.GOOS == "windows" && strings.ToLower(sourceDir) != strings.ToLower(destinationDir) && IsDirectoryExists(destinationDir) {
		return errors.New(fmt.Sprintf("Cannot move '%s' to the destination location since '%s' already exists.", sourceDir, destinationDir))
	}
	now := time.Now()
	uniqueId := fmt.Sprintf("%d", now.Unix())
	uniqueDirName := normalizedDestinationDirectory + "_" + uniqueId
	err := os.Rename(normalizedSourceDirectory, uniqueDirName)
	if err != nil {
		return err
	}
	err = os.Rename(uniqueDirName, normalizedDestinationDirectory)
	return err
}

/*
CreateDirectory allows you to create a directory on your local file system.
*/
func CreateDirectory(directoryPath string, permissions uint32) error {
	if permissions == 0 {
		permissions = 0744
	}
	perm := os.FileMode(permissions)
	err := os.MkdirAll(directoryPath, perm)
	return err
}

func CreateTemporaryFile(directoryPath string) (string, error) {
	// Create a temporary file in the specified directory
	tempFile, err := os.CreateTemp(directoryPath, "temp_") // "temp_" is the prefix for the file name
	if err != nil {
		return "", fmt.Errorf("could not create temporary file: %s", err.Error())
	}
	// Close the file immediately since we just need the file name
	defer tempFile.Close()

	// Return the file name (temporary file's name)
	return tempFile.Name(), nil
}

/*
GetFileExtension allows you to get the file extension of a file.
*/
func GetFileExtension(fileName string) string {
	extension := filepath.Ext(fileName)
	return extension
}

/*
GetBaseFileName allows you to extract the base name of a file without any path or
file extensions.
*/
func GetBaseFileName(fileName string) string {
	bareFileName := GetBareDirectoryPath(fileName)
	normalizedFileName := GetFileNameFromPath(bareFileName)
	return strings.TrimSuffix(normalizedFileName, filepath.Ext(normalizedFileName))
}

/*
GetBaseDirectory allows you to extract the directory from a file path.
*/
func GetBaseDirectory(filePath string) string {
	bareFilePath := GetBareDirectoryPath(filePath)
	directory, _ := filepath.Split(bareFilePath)
	return directory
}

/*
GetCurrentDirectory allows you to obtain the current directory from a fully
qualified directory path.
*/
func GetCurrentDirectory(directoryPath string) string {
	basePath := GetBaseDirectory(directoryPath)
	currentDirectory := strings.ReplaceAll(directoryPath, basePath, "")
	return currentDirectory
}

/*
GetFileSizeInKilobytes allows you to get the size of a file in kilobytes.
*/
func GetFileSizeInKilobytes(filePath string) (float64, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	fileSizeKB := float64(fileInfo.Size()) / 1024
	return fileSizeKB, nil
}

/*
GetFileNameFromPath allows you to extract a file name from a fully qualified file path.
*/
func GetFileNameFromPath(fullyQualifiedFileName string) string {
	return filepath.Base(fullyQualifiedFileName)
}

/*
*
IsFile allows you to check if an item on the file system is a file or not.
*/
func IsFile(path string) (bool, error) {
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

/*
RunCommandWithoutBlocking allows you to run a command without blocking execution. This is accomplished by returning
a pointer to the command output as well as a channel to signify when execution has completed.
This function properly handles UTF-8 encoded text in command output.
*/
func RunCommandWithoutBlocking(commandLineArguments ...string) (*string, <-chan struct{}, <-chan error) {
	cmd := exec.Command(commandLineArguments[0], commandLineArguments[1:]...)
	var cmdOutput bytes.Buffer
	cmd.Stdout = &cmdOutput
	cmd.Stderr = &cmdOutput
	var outputBuffer bytes.Buffer
	isExecutionDoneChannel := make(chan struct{})
	errorChannel := make(chan error, 1) // Buffered channel to avoid goroutine leak
	if err := cmd.Start(); err != nil {
		close(isExecutionDoneChannel)
		errorChannel <- err // Send error to the channel
		return nil, nil, errorChannel
	}
	go func() {
		if err := cmd.Wait(); err != nil {
			close(isExecutionDoneChannel)
			errorChannel <- err // Send error to the channel
			return
		}
		close(isExecutionDoneChannel)
	}()

	// Use a mutex to protect concurrent access to the output buffer
	var outputMutex sync.Mutex

	go func() {
		// Use a reader that preserves the original bytes including line endings
		reader := bufio.NewReader(&cmdOutput)
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil && err != io.EOF {
				break
			}

			if len(line) > 0 {
				outputMutex.Lock()
				outputBuffer.Write(line)
				outputMutex.Unlock()
			}

			if err == io.EOF {
				break
			}
		}
	}()

	// Wait for command to finish
	go func() {
		<-isExecutionDoneChannel
		// If the command exits with a non-zero status, return an error
		if cmd.ProcessState != nil && !cmd.ProcessState.Success() {
			errorChannel <- fmt.Errorf("%d", cmd.ProcessState.ExitCode())
		}
	}()

	// Convert buffer to string and return a pointer to it
	commandOutput := outputBuffer.String()
	return &commandOutput, isExecutionDoneChannel, errorChannel
}

func RunCommand(commandLineArguments ...string) (string, error) {
	cmd := exec.Command(commandLineArguments[0], commandLineArguments[1:]...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		errMsg := stderr.String()
		if errMsg == "" {
			errMsg = err.Error()
		}
		return "", errors.New(errMsg)
	}
	return stdout.String(), nil
}
