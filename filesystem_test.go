package filesystem

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"strings"
	"testing"
)

func TestGetCurrentDirectory(test *testing.T) {
	directory := "/tmp/directory1/directory2/This is a directory with spaces and . seperation"
	obtainedResult := GetCurrentDirectory(directory)
	expectedResult := "This is a directory with spaces and . seperation"
	assert.Equalf(test, expectedResult, obtainedResult, "The current directory was not as expected.")
}

func TestIsFileContainsText(test *testing.T) {
	fileName := "/tmp/testDocument.txt"
	if IsFileExists(fileName) {
		DeleteFile(fileName)
	}
	AppendLineToFile(fileName, "This is a first test line\n", 0)
	AppendLineToFile(fileName, "This is a second test line\n", 0)
	AppendLineToFile(fileName, "This is a third test line\n", 0)
	AppendLineToFile(fileName, "This is a -forth- test line\n", 0)
	expectedResult := true
	obtainedResult, _ := IsFileContainsText(fileName,".*second.*")
	assert.Equalf(test, expectedResult, obtainedResult, "The word 'second' was expected to be found in the file.")
}

func TestFindReplaceInFile(test *testing.T) {
	fileName := "/tmp/testDocument.txt"
	if IsFileExists(fileName) {
		DeleteFile(fileName)
	}
	AppendLineToFile(fileName, "This is a first test line\n", 0)
	AppendLineToFile(fileName, "This is a second test line\n", 0)
	AppendLineToFile(fileName, "This is a third test line\n", 0)
	AppendLineToFile(fileName, "This is a -forth- test line\n", 0)
	FindReplaceInFile(fileName, "-.*-", "REPLACED")
	FindReplaceInFile(fileName, ".*second.*", "DELETED")
	fileContents, err := GetFileContents(fileName)
	if err != nil {
		assert.NoErrorf(test, err, "An error was not expected when getting the contents of a file.")
	}
	expectedResult := true
	obtainedResult := strings.Contains(string(fileContents), "REPLACED")
	assert.Equalf(test, expectedResult, obtainedResult, "The word REPLACED was expected to be found in the file.")
	expectedResult = true
	obtainedResult = strings.Contains(string(fileContents), "DELETED")
	assert.Equalf(test, expectedResult, obtainedResult, "The word DELETED was expected to be found in the file.")
	expectedResult = false
	obtainedResult = strings.Contains(string(fileContents), "second")
	assert.Equalf(test, expectedResult, obtainedResult, "The word second was not expected to be found in the file.")
	expectedResult = false
	obtainedResult = strings.Contains(string(fileContents), "-forth-")
	assert.Equalf(test, expectedResult, obtainedResult, "The word -forth- was not expected to be found in the file.")

}

func TestIsDirectory(test *testing.T) {
	path := "/tmp/"
	isDirectory := IsDirectory(path)
	assert.Equalf(test, true, isDirectory, "The path provided should return as a directory.")
	path = "/tmp/asdasds"
	isDirectory = IsDirectory(path)
	assert.Equalf(test, false, isDirectory, "The path provided should return as a directory.")
}

func TestDeleteDirectory(test *testing.T) {
	directory := "/tmp/dir_test"
	subDirectory := "/tmp/dir_test/sub_dir"
	err := CreateDirectory(directory, 0)
	if err != nil {
		assert.NoErrorf(test, err, "An error was not expected when creating a directory.")
	}
	err = CreateDirectory(subDirectory, 0)
	if err != nil {
		assert.NoErrorf(test, err, "An error was not expected when creating a sub directory.")
	}
	obtainedResult := IsDirectoryExists(directory)
	expectedResult := true
	assert.Equalf(test,expectedResult, obtainedResult, "The created directory was expected to exist.")
	obtainedResult = IsDirectoryExists(subDirectory)
	expectedResult = true
	assert.Equalf(test,expectedResult, obtainedResult, "The created sub directory was expected to exist.")
	err = DeleteDirectory(directory)
	if err != nil {
		assert.NoErrorf(test, err, "An error was not expected when deleting a directory")
	}
	obtainedResult = IsDirectoryExists(directory)
	expectedResult = false
	assert.Equalf(test,expectedResult, obtainedResult, "The delete directory was not expected to exist.")
	obtainedResult = IsDirectoryExists(subDirectory)
	expectedResult = false
	assert.Equalf(test,expectedResult, obtainedResult, "The deleted directory was not expected to exist.")
}

func TestDownloadFile(test *testing.T) {
	err := DownloadFile("https://bad_url", "/tmp/download.txt", nil)
	assert.Errorf(test, err, "An error was expected to be generated when a bad URL is used!")
	err = DownloadFile("https://www.google.ca", "/tmp/index.html", nil)
	assert.NoErrorf(test, err, "An error was not expected to be generated when downloading!")
	err = DownloadFile("https://www.google.ca", "/tmp/index.html", nil)
	assert.NoErrorf(test, err, "An error was not expected to be generated when downloading!")
	var header http.Header
	header = make(http.Header)
	header.Set("User-Agent", "Mozilla/5.0 (X11; Fedora; Linux x86_64; rv:52.0) Gecko/20100101 Firefox/52.0")
	header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	err = DownloadFile("https://www.google.ca", "/tmp/index.html", header)
	assert.NoErrorf(test, err, "An error was not expected to be generated when downloading!")
}

func TestGetLinesFromFile(test *testing.T) {
	var file fileInstanceType
	filename := "/tmp/file.txt"
	if IsFileExists(filename) {
		DeleteFile(filename)
	}
	err := file.Open(filename, 0)
	if err != nil {
		assert.NoErrorf(test, err, "An error was not expected when opening a file.")
	}
	err = file.WriteLine("First written line.")
	if err != nil {
		assert.NoErrorf(test, err, "An error was not expected when writing lines to a file.")
	}
	err = file.WriteLine("Second written line.")
	if err != nil {
		assert.NoErrorf(test, err, "An error was not expected when writing lines to a file.")
	}
	err = file.WriteLine("Third written line.")
	if err != nil {
		assert.NoErrorf(test, err, "An error was not expected when writing lines to a file.")
	}
	file.Close()
	fileContents, err := GetFileContents(filename)
	if err != nil {
		assert.NoErrorf(test, err, "An error was not expected when getting lines from a file.")
	}
	assert.Equalf(test,"First written line.\nSecond written line.\nThird written line.\n", string(fileContents), "The text file was expected to be a size it wasn't.")
}

func TestGetLastLineFromFile(test *testing.T) {
	var file fileInstanceType
	filename := "/tmp/file.txt"
	if IsFileExists(filename) {
		DeleteFile(filename)
	}
	err := file.Open(filename, 0)
	if err != nil {
		assert.NoErrorf(test, err, "An error was not expected when opening a file.")
	}
	err = file.WriteLine("First written line.")
	if err != nil {
		assert.NoErrorf(test, err, "An error was not expected when writing lines to a file.")
	}
	err = file.WriteLine("Second written line.")
	if err != nil {
		assert.NoErrorf(test, err, "An error was not expected when writing lines to a file.")
	}
	err = file.WriteLine("Third written line.")
	if err != nil {
		assert.NoErrorf(test, err, "An error was not expected when writing lines to a file.")
	}
	file.Close()
	fileContents, err := GetLastLineFromFile(filename)
	if err != nil {
		assert.NoErrorf(test, err, "An error was not expected when getting lines from a file.")
	}
	assert.Equalf(test,"Third written line.", string(fileContents), "The text file was expected to be a size it wasn't.")
}

func TestRemoveFirstLineFromFile(test *testing.T) {
	filename := "/tmp/file.txt"
	if IsFileExists(filename) {
		DeleteFile(filename)
	}
	err := AppendLineToFile(filename, "First written line.\n", 0)
	if err != nil {
		assert.NoErrorf(test, err, "An error occured while trying to append the first line.")
	}
	err = AppendLineToFile(filename, "Second written line.\n", 0)
	if err != nil {
		assert.NoErrorf(test, err, "An error occured while trying to append the second line.")
	}
	err = AppendLineToFile(filename, "Third written line.", 0)
	if err != nil {
		assert.NoErrorf(test, err, "An error occured while trying to append the third line.")
	}
	fileContents, err := GetFileContents(filename)
	if err != nil {
		assert.NoErrorf(test, err, "An error was not expected when reading a file.")
	}
	assert.Equalf(test, "First written line.\nSecond written line.\nThird written line.", string(fileContents), "The text file was expected to be a size it wasn't.")
	err = RemoveFirstLineFromFile(filename)
	if err != nil {
		assert.NoErrorf(test, err, "An error was not expected when trying to remove the first line of a file.")
	}
	fileContents, err = GetFileContents(filename)
	if err != nil {
		assert.NoErrorf(test, err, "An error was not expected when reading a file.")
	}
	assert.Equalf(test, "Second written line.\nThird written line.", string(fileContents), "The text file was expected to be a size it wasn't.")
}

func TestGetFileContents(test *testing.T) {
	var file fileInstanceType
	filename := "/tmp/file.txt"
	if IsFileExists(filename) {
		DeleteFile(filename)
	}
	err := file.Open(filename, 0)
	if err != nil {
		assert.NoErrorf(test, err, "An error was not expected when opening a file.")
	}
	err = file.WriteLine("First written line.")
	if err != nil {
		assert.NoErrorf(test, err, "An error was not expected when writing lines to a file.")
	}
	err = file.WriteLine("Second written line.")
	if err != nil {
		assert.NoErrorf(test, err, "An error was not expected when writing lines to a file.")
	}
	err = file.WriteLine("Third written line.")
	if err != nil {
		assert.NoErrorf(test, err, "An error was not expected when writing lines to a file.")
	}
	file.Close()
	err = file.Open(filename, 0)
	if err != nil {
		assert.NoErrorf(test, err, "An error was not expected when opening a file.")
	}
	fileContents, err := file.GetFileContents()
	if err != nil {
		assert.NoErrorf(test, err, "An error was not expected when reading a file.")
	}
	assert.Equalf(test, "First written line.\nSecond written line.\nThird written line.", string(fileContents), "The text file was expected to be a size it wasn't.")
}
func TestPopLineFromStack(test *testing.T) {
	var file fileInstanceType
	filename := "/tmp/file.txt"
	if IsFileExists(filename) {
		DeleteFile(filename)
	}
	err := file.Open(filename, 0)
	if err != nil {
		assert.NoErrorf(test, err, "An error was not expected when opening a file.")
	}
	err = file.WriteLine("First written line.")
	if err != nil {
		assert.NoErrorf(test, err, "An error was not expected when writing lines to a file.")
	}
	err = file.WriteLine("Second written line.")
	if err != nil {
		assert.NoErrorf(test, err, "An error was not expected when writing lines to a file.")
	}
	err = file.WriteLine("Third written line.")
	if err != nil {
		assert.NoErrorf(test, err, "An error was not expected when writing lines to a file.")
	}
	file.Close()
	err = file.Open(filename, 0)
	if err != nil {
		assert.NoErrorf(test, err, "An error was not expected when opening a file.")
	}
	file.RemoveFirstLine()
	file.Close()
	fileSize, _:= GetFileSize(filename)
	assert.Equalf(test, int64(41), fileSize, "The text file was expected to be a size it wasn't.")
	err = file.Open(filename, 0)
	if err != nil {
		assert.NoErrorf(test, err, "An error was not expected when opening a file.")
	}
	obtainedResult, err := file.GetFirstLine()
	if err != nil {
		assert.NoErrorf(test, err, "An error was not expected obtaining the first line of a file.")
	}
	assert.Equal(test, "Second written line.", string(obtainedResult), "The first line read was not as expected.")
}

func TestFileOperations(test *testing.T) {
	var file fileInstanceType
	filename := "/tmp/file.txt"
	if IsFileExists(filename) {
		DeleteFile(filename)
	}
	err := file.Open(filename, 0)
	if err != nil {
		assert.NoErrorf(test, err, "An error was not expected when opening a file.")
	}
	err = file.WriteBytes([]byte{255,255,255})
	if err != nil {
		assert.NoErrorf(test, err, "An error was not expected when writing bytes to a file.")
	}
	err = file.WriteString("\nNew line\n")
	if err != nil {
		assert.NoErrorf(test, err, "An error was not expected when writing a string to a file.")
	}
	file.Close()
	assert.True(test, IsFileExists(filename), "A text file was expected to exist when it doesn't.")
	fileSize, _:= GetFileSize(filename)
	assert.Equalf(test, int64(13), fileSize, "The text file was expected to be a size it wasn't.")
}

func TestCopyFile(test *testing.T) {
	sourceFile := "/tmp/source.txt"
	targetFile := "/tmp/target.txt"
	err := WriteBytesToFile(sourceFile, []byte("sample_string"), 0666)
	assert.NoErrorf(test, err, "An error was not expected when creating a sample file!")
	err = CopyFile(sourceFile, targetFile)
	assert.NoErrorf(test, err, "An error was not expected when copying a sample file!")
	err = DeleteFile(sourceFile)
	assert.NoErrorf(test, err, "An error was not expected when trying to delete a source sample file!")
	err = DeleteFile(targetFile)
	assert.NoErrorf(test, err, "An error was not expected when trying to delete a target sample file!")
}

func TestAppendLinesToFile(test *testing.T) {
	sourceFile := "/tmp/source.txt"
	err := WriteBytesToFile(sourceFile, []byte("sample_string"), 0666)
	assert.NoErrorf(test, err, "An error was not expected when creating a sample file!")
	err = AppendLineToFile(sourceFile, "A sample line", 0666)
	assert.NoErrorf(test, err, "An error was not expected when trying to append to a file!")
	err = DeleteFile(sourceFile)
	assert.NoErrorf(test, err, "An error was not expected when trying to delete a source sample file!")
}

func TestGetListOfFiles(test *testing.T) {
	sourceDirectory := "/tmp/"
	_, err := GetListOfFiles(sourceDirectory, ".*")
	assert.NoErrorf(test, err, "An error was not expected when trying to list all files within a directory!")
}

func TestGetListOfDirectories(test *testing.T) {
	sourceDirectory := "/tmp/"
	_, err := GetListOfDirectories(sourceDirectory, ".*")
	assert.NoErrorf(test, err, "An error was not expected when trying to list all directories within a directory!")
}

func TestGetNormalizedDirectoryPath(test *testing.T) {
	sourceDirectory := "/tmp/someDirectory"
	obtainedValue := GetNormalizedDirectoryPath(sourceDirectory)
	expectedValue := sourceDirectory + "/"
	assert.Equalf(test, expectedValue, obtainedValue, "The directory path was not normalized as expected!")
}

func TestGetDefaultCacheDirectory(test *testing.T) {
	_, err := GetDefaultCacheDirectory()
	assert.NoErrorf(test, err, "An error was not expected when trying to detect the default cache directory!")
}

func TestFindMatchingContent(test *testing.T) {
	sourceDirectory := "/tmp/"
	_, err := FindMatchingContent(sourceDirectory, []string{"should_never_match", ".*"}, false, true, true)
	assert.NoErrorf(test, err, "An error was not expected when trying to search the contents of a directory!")

	_, err = FindMatchingContent(sourceDirectory, []string{"should_never_match",".*"}, true, false, false)
	assert.NoErrorf(test, err, "An error was not expected when trying to search the contents of a directory!")
}

func TestGetListOfDirectoryContents(test *testing.T) {
	sourceDirectory := "/tmp/"
	_, err := GetListOfDirectoryContents(sourceDirectory, []string{".*"}, true, true)
	assert.NoErrorf(test, err, "An error was not expected when trying to list the contents of a directory!")
}

func TestIsDirectoryEmpty(test *testing.T) {
	sourceDirectory := "/tmp/"
	sourceFile := "/tmp/source.txt"
	err := WriteBytesToFile(sourceFile, []byte("sample_string"), 0666)
	assert.NoErrorf(test, err, "An error was not expected when creating a sample file!")
	obtainedValue, err := IsDirectoryEmpty(sourceDirectory)
	expectedValue := false
	assert.NoErrorf(test, err, "An error was not expected when trying check if a directory was empty or not!")
	assert.Equalf(test, expectedValue, obtainedValue, "The directory specified was expected to be not empty!")
	err = DeleteFile(sourceFile)
	assert.NoErrorf(test, err, "An error was not expected when trying to delete a source sample file!")
}

func TestGetWorkingDirectory(test *testing.T) {
	_, err := GetWorkingDirectory()
	assert.NoErrorf(test, err, "An error was not expected when trying to get the current working directory!")
}

func TestGetParentDirectory(test *testing.T) {
	obtainedValue := GetParentDirectory("/tmp/some_directory/sample_file.txt")
	expectedValue := "/tmp/some_directory"
	assert.Equalf(test, expectedValue, obtainedValue, "The parent directory did not match what was expected!")
}

func TestRenameFile(test *testing.T) {
	sourceFile := "/tmp/source.txt"
	targetFile := "/tmp/target.txt"
	err := WriteBytesToFile(sourceFile, []byte("sample_string"), 0666)
	assert.NoErrorf(test, err, "An error was not expected when creating a sample file!")
	err = RenameFile(sourceFile, targetFile)
	assert.NoErrorf(test, err, "No error was expected when trying to rename a file!")
	err = DeleteFile(targetFile)
	assert.NoErrorf(test, err, "An error was not expected when trying to delete a source sample file!")
}

func TestIsDirectoryExists(test *testing.T) {
	obtainedValue := IsDirectoryExists("/tmp")
	expectedValue := true
	assert.Equalf(test, expectedValue, obtainedValue, "The specified directory was expected to exist!")
}

func TestDeleteFile(test *testing.T) {
	sourceFile := "/tmp/source.txt"
	err := WriteBytesToFile(sourceFile, []byte("sample_string"), 0666)
	assert.NoErrorf(test, err, "An error was not expected when creating a sample file!")
	err = DeleteFile(sourceFile)
	assert.NoErrorf(test, err, "An error was not expected when trying to delete a source sample file!")
}

func TestMoveFile(test *testing.T) {
	sourceFile := "/tmp/source.txt"
	targetFile := "/tmp/target.txt"
	err := WriteBytesToFile(sourceFile, []byte("sample_string"), 0666)
	assert.NoErrorf(test, err, "An error was not expected when creating a sample file!")
	err = MoveFile(sourceFile, targetFile)
	assert.NoErrorf(test, err, "An error was not expected when copying a sample file!")
	obtainedValue := IsDirectoryExists(sourceFile)
	expectedValue := false
	assert.Equalf(test, expectedValue, obtainedValue, "The source directory that was moved was not expected to exist!")
	err = DeleteFile(targetFile)
	assert.NoErrorf(test, err, "An error was not expected when trying to delete a target sample file!")
}

func TestGetFileContentsAsBytes(test *testing.T) {
	sourceFile := "/tmp/source.txt"
	expectedValue := []byte("This is a sample string!")
	err := WriteBytesToFile(sourceFile, expectedValue, 0666)
	assert.NoErrorf(test, err, "An error was not expected when creating a sample file!")
	obtainedValue, err := GetFileContentsAsBytes(sourceFile)
	assert.Equalf(test, expectedValue, obtainedValue, "The obtained value did not match what was expected!")
	err = DeleteFile(sourceFile)
	assert.NoErrorf(test, err, "An error was not expected when trying to delete a target sample file!")
}

func TestDeleteFilesMatchingPattern(test *testing.T) {
	err := WriteBytesToFile("/tmp/file1.txt", []byte("sample_string"), 0666)
	err = WriteBytesToFile("/tmp/file2.txt", []byte("sample_string"), 0666)
	err = WriteBytesToFile("/tmp/file3.txt", []byte("sample_string"), 0666)
	err = DeleteFilesMatchingPattern("/tmp/file*.txt")
	assert.NoErrorf(test, err, "An error was not expected when trying to delete a target sample file!")
}

func TestGetFileSize(test *testing.T) {
	sourceFile := "/tmp/source.txt"
	fileContents := []byte("This is a sample string!")
	err := WriteBytesToFile(sourceFile, fileContents, 0666)
	assert.NoErrorf(test, err, "An error was not expected when creating a sample file!")
	obtainedValue, err := GetFileSize(sourceFile)
	expectedValue := int64(24)
	assert.Equalf(test, expectedValue, obtainedValue, "The obtained file size did not match what was expected.")
	assert.NoErrorf(test, err, "An error was not expected when trying to get the size of the file!")
	err = DeleteFile(sourceFile)
	assert.NoErrorf(test, err, "An error was not expected when trying to delete a target sample file!")
}


func TestCreateDirectory(test *testing.T) {
	sourceDirectory := "/tmp/newDirectory"
	err := CreateDirectory(sourceDirectory, 0666)
	assert.NoErrorf(test, err, "No error was expected to occur when trying to create a directory!")
	err = DeleteFile(sourceDirectory)
	assert.NoErrorf(test, err, "No error was expected to occur when trying to delete a directory!")
}

func TestGetFileExtension(test *testing.T) {
	sourceFile := "/tmp/source.txt"
	fileContents := []byte("This is a sample string!")
	err := WriteBytesToFile(sourceFile, fileContents, 0666)
	assert.NoErrorf(test, err, "An error was not expected when creating a sample file!")
	obtainedValue := GetFileExtension(sourceFile)
	expectedValue := ".txt"
	assert.Equalf(test, expectedValue, obtainedValue, "The file extension obtained was not what was expected!")
	err = DeleteFile(sourceFile)
	assert.NoErrorf(test, err, "No error was expected to occur when trying to delete a file!")
}

func TestGetBaseFileName(test *testing.T) {
	obtainedValue := GetBaseFileName("my_file.eng.txt")
	expectedValue := "my_file.eng"
	assert.Equalf(test, expectedValue, obtainedValue, "The base filename obtained was not what was expected!")
}

func TestGetFileNameFromPath(test *testing.T) {
	sourceFile := "/tmp/source.txt"
	fileContents := []byte("This is a sample string!")
	err := WriteBytesToFile(sourceFile, fileContents, 0666)
	assert.NoErrorf(test, err, "An error was not expected when creating a sample file!")
	obtainedValue := GetFileNameFromPath(sourceFile)
	expectedValue := "source.txt"
	assert.Equalf(test, expectedValue, obtainedValue, "The file name obtained from the provided path was not what was expected!")
	err = WriteBytesToFile(sourceFile, fileContents, 0666)
	assert.NoErrorf(test, err, "An error was not expected when creating a sample file!")
}

func TestGetBaseDirectory(test *testing.T) {
	sourceFile := "/tmp/source.txt"
	fileContents := []byte("This is a sample string!")
	err := WriteBytesToFile(sourceFile, fileContents, 0666)
	assert.NoErrorf(test, err, "An error was not expected when creating a sample file!")
	obtainedValue := GetBaseDirectory(sourceFile)
	expectedValue := "/tmp/"
	assert.Equalf(test, expectedValue, obtainedValue, "The base directory obtained was not what was expected!")
	err = WriteBytesToFile(sourceFile, fileContents, 0666)
	assert.NoErrorf(test, err, "An error was not expected when creating a sample file!")
}

func TestIsFile(test *testing.T) {
	sourceDirectory := "/tmp/newDirectory"
	err := CreateDirectory(sourceDirectory, 0666)
	assert.NoErrorf(test, err, "No error was expected to occur when trying to create a directory!")
	obtainedValue, err := IsFile(sourceDirectory)
	expectedValue := false
	assert.NoErrorf(test, err, "An error was not expected when checking if a path is a directory or not!")
	assert.Equalf(test, expectedValue, obtainedValue, "The value returned for checking if a path was a file or not was not as expected!")
	err = DeleteFile(sourceDirectory)
	assert.NoErrorf(test, err, "No error was expected to occur when trying to delete a directory!")
}