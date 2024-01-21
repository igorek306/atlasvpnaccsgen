# AtlasVPN accounts generator
## Purpose

Purpose of that project is to generate AtlasVPN accounts with 70 days premium for **free.**

## How it works

First, the app creates account - this one you will be using and downloads its uuid.\
Next asynchronously creates 10 accounts using uuid as referrer code.\
Why only 10? Because only for 10 invited 'friends' you will get premium days.

# Installation 
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