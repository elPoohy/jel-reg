# jel-reg
Средство для создания пользователей в Jelastic.

Всё жёстко захардкоженно так как писалось за минуту

Новые порталы добавлюятся в двух точках в index.html в разделе

### В index.html

```html
<select name="portal" class="custom-select" id="inputGroupSelect02">
      <option value="DataFort" {{ if eq .Portal "DataFort" }} selected {{ end }}>Клиент DataFort</option>
      <option value="Beeline" {{ if eq .Portal "Beeline"  }} selected {{ end }}>Клиент Beeline</option>
      <option value="SysSoft" {{ if eq .Portal "SysSoft"  }} selected {{ end }}>Клиент SysSoft</option>
</select>
```
### В main.go

```go
switch details.Portal {
  case "DataFort":
    apiURL = "https://reg.paasinfra.datafort.ru"
  case "Beeline":
    apiURL = "https://reg.paas.beelinecloud.ru/"
  case "SysSoft":
    apiURL = "https://reg.paas.syssoft.ru"
}
```
