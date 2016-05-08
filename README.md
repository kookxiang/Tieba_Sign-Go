# Tieba Sign in Go

**WARNING: This project is currently under BETA, use it on your risk**

A faster && lighter sign robot for http://tieba.baidu.com which can help you get more credit at tieba.

## Usage

Notice: if you want to sign in automatically, please add it to crontab list (for Linux) or Task Schedule (for Windows) to run at every morning.

## Single Mode / User Interactive Mode

Just execute the compiled program :D

Input username and password to login if you use this program at first time.

If verification code is needed, a file named `captcha.jpg` will be downloaded to current directory, and you need to open and type it to the program.

If you cannot login via user interactive mode, you need to gather cookie information by yourself. Put a file named `cookie.txt` in current directory with the following format:

```
BDUSS=YOUR-BDUSS-SECRET-CODE-IN-COOKIE
(Others is optional)
```

Once you'd logged in, a `cookie.txt` file will generated automatically. And next time you run it, you will not need to login again.

### Multi-User Mode / Batch Mode

Just create a directory named `cookies` and copy the `cookie.txt` file to the new `cookies` directory, you can change the file name on your own (but please keep the `.txt` extensions).

Now you can put more `.txt` named cookie files into `cookies` folder.

Execute the program with parameter `-batch`, the program will handle all user by the same time.

If you execute this little program by batch mode, it will create go routines for every user and sign it in same time. (Like multi-thread, but more effective. [Learn More](https://golang.org/doc/effective_go.html#goroutines))

## Installation

You can both compile this program by yourself of just download the prebuilt version.

If you don't know how to compile it, please [Click here to DOWNLOAD](https://github.com/kookxiang/Tieba_Sign-Go/releases) the prebuilt version.

### Compile

1. First, install go environment on your system and configure GOPATH.

2. Install dependency:

   ```shell
   go get github.com/bitly/go-simplejson
   go get golang.org/x/text/encoding
   go get golang.org/x/text/encoding/simplifiedchinese
   go get golang.org/x/text/transform
   ```

3. Compile project

   ```shell
   go build
   ```

4. You're done!