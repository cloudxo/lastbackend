package handlers

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/deployithq/deployit/env"
	"github.com/deployithq/deployit/errors"
	"github.com/deployithq/deployit/utils"
	"github.com/fatih/color"
	"gopkg.in/urfave/cli.v2"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// Main deploy it handler
// - Archive all files which is in folder
// - Send it to server

func It(c *cli.Context) error {

	env := NewEnv()

	var archiveName string = "tar.gz"
	var archivePath string = fmt.Sprintf("%s/.dit/%s", env.Path, archiveName)

	appInfo := new(AppInfo)
	err := appInfo.Read(env.Log, env.Path, env.Host)
	if err != nil {
		env.Log.Error(err)
		return err
	}

	if appInfo.Name == "" {

		appInfo.Name = utils.AppName(env.Path)
		appInfo.Tag = Tag

		appInfo.UUID, err = AppCreate(env, appInfo.Name, appInfo.Tag)
		if err != nil {
			env.Log.Error(err)
			return err
		}

		color.Cyan("Creating app: %s", appInfo.Name)
	} else {
		color.Cyan("Updating app: %s", appInfo.Name)
	}

	// Creating archive
	fw, err := os.Create(archivePath)
	if err != nil {
		env.Log.Error(err)
		return err
	}

	gw := gzip.NewWriter(fw)
	tw := tar.NewWriter(gw)

	// Deleting archive after function ends
	defer func() {
		env.Log.Debug("Deleting archive: ", archivePath)

		fw.Close()
		gw.Close()
		tw.Close()

		// Deleting files
		err = os.Remove(archivePath)
		if err != nil {
			env.Log.Error(err)
			return
		}
	}()

	// Listing all files from database to know what files were deleted from previous run
	storedFiles, err := env.Storage.ListAllFiles(env.Log)
	if err != nil {
		env.Log.Error(err)
		return err
	}

	// TODO Include deleted folders to deletedFiles like "nginx/"

	excludePatterns, err := utils.LoadDockerPatterns(env.Path)
	if err != nil {
		env.Log.Error(err)
		return err
	}

	excludePatterns = append(excludePatterns, ".gitignore", ".dit", ".git")

	color.Cyan("Packing files")
	storedFiles, err = PackFiles(env, tw, env.Path, storedFiles, excludePatterns)
	if err != nil {
		return err
	}

	deletedFiles := []string{}

	for key, _ := range storedFiles {
		env.Log.Debug("Deleting: ", key)
		err = env.Storage.Delete(env.Log, key)
		if err != nil {
			return err
		}

		deletedFiles = append(deletedFiles, key)
	}

	tw.Close()
	gw.Close()
	fw.Close()

	bodyBuffer := new(bytes.Buffer)
	bodyWriter := multipart.NewWriter(bodyBuffer)

	// Adding deleted files to request
	if len(deletedFiles) > 0 {
		delFiles, err := json.Marshal(deletedFiles)
		if err != nil {
			env.Log.Error(err)
			return err
		}

		bodyWriter.WriteField("deleted", string(delFiles))
	}

	archiveInfo, err := os.Stat(archivePath)
	if err != nil {
		env.Log.Error(err)
		return err
	}

	// If archive size is 32 it means that it is empty and we don't need to send it
	if archiveInfo.Size() != 32 {
		fh, err := os.Open(archivePath)
		if err != nil {
			env.Log.Error(err)
			return err
		}

		fileWriter, err := bodyWriter.CreateFormFile("file", "tar.gz")
		if err != nil {
			env.Log.Error(err)
			return err
		}

		_, err = io.Copy(fileWriter, fh)
		if err != nil {
			env.Log.Error(err)
			return err
		}

		fh.Close()
	}

	bodyWriter.Close()

	// TODO If error in response: rollback hash table

	// Creating response for file uploading with fields
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/app/%s/deploy", env.HostUrl, appInfo.UUID), bodyBuffer)
	if err != nil {
		env.Log.Error(err)
		return err
	}

	req.Header.Set("Content-Type", bodyWriter.FormDataContentType())

	color.Cyan("Uploading sources")

	client := new(http.Client)
	res, err := client.Do(req)
	if err != nil {
		env.Log.Error(err)
		return err
	}

	if Log {
		color.Cyan("Logs: ")
		reader := bufio.NewReader(res.Body)
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				break
			}
			fmt.Println(string(line))
		}
	}

	err = appInfo.Write(env.Log, env.Path, env.Host, appInfo.UUID, appInfo.Name, appInfo.Tag, appInfo.URL)
	if err != nil {
		env.Log.Error(err)
		return err
	}

	res.Body.Close()

	color.Cyan("Done")

	return nil
}

func PackFiles(env *env.Env, tw *tar.Writer, filesPath string, storedFiles map[string]string, excludePatterns []string) (map[string]string, error) {

	// Opening directory with files
	dir, err := os.Open(filesPath)

	if err != nil {
		env.Log.Error(err)
		return storedFiles, err
	}

	// Reading all files
	files, err := dir.Readdir(0)
	if err != nil {
		env.Log.Error(err)
		return storedFiles, err
	}

	for _, file := range files {

		fileName := file.Name()

		currentFilePath := fmt.Sprintf("%s/%s", filesPath, fileName)

		// Creating path, which will be inside of archive
		relativePath := strings.Replace(currentFilePath, env.Path, "", 1)[1:]

		// Ignoring files which is not needed for build to make archive smaller
		// TODO: create base .ditignore file on first application creation

		matches, err := utils.Matches(relativePath, excludePatterns)
		if err != nil {
			return storedFiles, err
		}
		if matches {
			continue
		}

		// If it was directory - calling this function again
		// In other case adding file to archive
		if file.IsDir() {
			storedFiles, err = PackFiles(env, tw, currentFilePath, storedFiles, excludePatterns)
			if err != nil {
				return storedFiles, err
			}
			continue
		}

		// Creating hash
		hash := utils.Hash(fmt.Sprintf("%s:%s:%s", file.Name(), strconv.FormatInt(file.Size(), 10), file.ModTime()))

		if storedFiles[relativePath] == hash {
			delete(storedFiles, relativePath)
			continue
		}

		delete(storedFiles, relativePath)

		// If hashes are not equal - add file to archive
		env.Log.Debug("Packing file: ", currentFilePath)

		err = env.Storage.Write(env.Log, relativePath, hash)
		if err != nil {
			return storedFiles, err
		}

		fr, err := os.Open(currentFilePath)
		if err != nil {
			env.Log.Error(err)
			return storedFiles, err
		}

		h := &tar.Header{
			Name:    relativePath,
			Size:    file.Size(),
			Mode:    int64(file.Mode()),
			ModTime: file.ModTime(),
		}

		err = tw.WriteHeader(h)
		if err != nil {
			env.Log.Error(err)
			return storedFiles, err
		}

		_, err = io.Copy(tw, fr)
		if err != nil {
			env.Log.Error(err)
			return storedFiles, err
		}

		fr.Close()

	}

	dir.Close()

	return storedFiles, err

}

func AppCreate(env *env.Env, name, tag string) (string, error) {

	var uuid string

	request := struct {
		Name string `json:"name"`
		Tag  string `json:"tag"`
	}{name, tag}

	var buf io.ReadWriter
	buf = new(bytes.Buffer)

	err := json.NewEncoder(buf).Encode(request)
	if err != nil {
		return uuid, err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/app", env.HostUrl), buf)
	if err != nil {
		env.Log.Error(err)
		return uuid, err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	client := new(http.Client)
	res, err := client.Do(req)
	if err != nil {
		env.Log.Error(err)
		return uuid, err
	}

	if res.StatusCode != 200 {
		err = errors.ParseError(res)
		env.Log.Error(err)
		return uuid, err
	}

	response := struct {
		UUID string `json:"uuid"`
	}{}

	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		env.Log.Error(err)
		return uuid, err
	}

	uuid = response.UUID

	return uuid, err
}