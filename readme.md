# AtlasVPN accounts generator
## Purpose

Purpose of that project is to generate AtlasVPN accounts with 70 days premium for **free.**

## How it works

First, the app creates account - this one you will be using and downloads its uuid.\
Next asynchronously creates 10 accounts using uuid as referrer code.\
Why only 10? Because only for 10 invited 'friends' you will get premium days.

#  Installation & usage 
```bash
git clone https://github.com/igorek306/atlasvpnaccsgen
```
```bash
cd atlasvpnaccsgen
```
```bash
go mod tidy
```
```bash
go run .
```
Now wait until your account is registered then you will get email address hosted on [1secmail](https://www.1secmail.com/). After registration, until you close the app, it will be listening to mails from AtlasVPN and printing auth links. If you close the app, you will still have access to your mailbox via the [1secmail](https://www.1secmail.com/) website.

# ⭐ Features

- **unlimited** number of generations
- easy login - you don't have to check mailbox yourself, this app does it automatically and prints link
- you can always login again to you generated AtlasVPN account - app saves all created accounts in **logs.txt** file inside app directory

# ❗Note
After creating several accounts within a short period of time, you may encounter http errors such as **403 Forbidden** or **503 Service Unavailable**. 
To fix that you can **simply wait or change your IP** address using AtlasVPN you already got or proxies.

## ⚠️ DISCLAIMER
This github repositiory is for EDUCATIONAL PURPOSES ONLY. I am NOT under any responsibility if a problem occurs.
