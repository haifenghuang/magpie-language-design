/* 使用命令对象的方式的脚本代码 */
res = `curl.exe -s https://api.ipify.org?format=json`
if !res.ok() {
  printf("An error occurred: %s\n", res)
} else {
  printf("res: %s\n", res)
}

date = `date /t`
if !date.ok() {
  printf("An error occurred: %s\n", res)
} else {
  printf("date: %s\n", date)
}



/* 使用返回多个值的方式的脚本代码
res, ok = `curl.exe -s https://api.ipify.org?format=json`

if !ok {
  printf("An error occurred: %s\n", res)
} else {
  printf("res: %s\n", res)
}

res, ok = `date /t`
if !ok {
  printf("An error occurred: %s\n", res)
} else {
  printf("date: %s\n", res)
}
*/