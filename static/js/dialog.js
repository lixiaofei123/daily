
function validateWithRegex(regexStr, str) {
    var regex = new RegExp(regexStr);
    return regex.test(str);
}

function alertDialog(config, confirm, cancel) {

    confirm = confirm || async function () { }
    cancel = cancel || function () { }

    let dialog = document.createElement("div")
    dialog.className = "dialog"

    let innerDialog = document.createElement("div")
    innerDialog.className = "inner-dialog"

    if (config.title) {
        let title = document.createElement("div")
        title.className = "title"
        title.innerText = config.title
        innerDialog.appendChild(title)
    }

    let body = document.createElement("div")
    body.className = "body"

    if (config.type === 'message' || config.type === 'confirm') {
        body.innerHTML = config.text
    }

    if(config.type === 'image'){
        innerDialog.style.width = "90%"
        innerDialog.style.maxWidth = "600px"
        body.style.textAlign = "center"
        body.innerHTML = `<img style="max-width: 100%;max-height: 500px;" src="${config.url}" />`
    }

    if(config.type === 'video'){
        innerDialog.style.width = "90%"
        innerDialog.style.maxWidth = "600px"
        body.style.textAlign = "center"
        body.innerHTML = `<video style="max-width: 100%;max-height: 500px;" src="${config.url}" autoplay />`
    }

    let data = {}
    let validateResult = {}
    let formerror;

    if (config.type === 'form') {
        let items = config.items
        for (let i = 0; i < items.length; i++) {
            data[items[i].name] = ""
            let formtiem = document.createElement("div")
            formtiem.className = "formitem"
            let span = document.createElement("span")
            span.innerText = items[i].label
            formtiem.appendChild(span)

            let input = document.createElement("input")
            if(items[i].name === "password"){
                input.type = "password"
            }
            input.addEventListener("change", () => {
                data[items[i].name] = input.value
            })

            formtiem.appendChild(input)
            body.appendChild(formtiem)
        }
        formerror = document.createElement("div")
        formerror.className = "formerror"
        formerror.style.display = "none"
        body.appendChild(formerror)

        if(config.tip){
            tip = document.createElement("div")
            tip.className = "tips"
            tip.innerText = config.tip
            body.appendChild(tip)
        }
    }

    if(config.type === "options"){
        let items = config.items
        let menu = document.createElement("ul")
        for (let i = 0; i < items.length; i++) {
            let item = items[i]
            let menuitem = document.createElement("li")
            menuitem.innerText = item.label
            menuitem.value = item.value
            menuitem.addEventListener("click", ()=>{
                confirm(item)
                cancelButton.click()
            })
            menu.appendChild(menuitem)
        }
        body.appendChild(menu)
    }

    innerDialog.appendChild(body)

    let bottom = document.createElement("div")
    bottom.className = "bottom"


    innerDialog.appendChild(bottom)

    let confirmButton = document.createElement("button")
    confirmButton.innerText = "确认"
    confirmButton.className = "confirm"


    let cancelButton = document.createElement("button")
    cancelButton.innerText = "关闭"
    cancelButton.className = "cancel"
    cancelButton.addEventListener("click", () => {
        cancel()
        dialog.remove()
    })


    if (config.type === 'confirm') {
        if (config.subtype === 'danger') {
            confirmButton.className = "danger"
        } else {
            confirmButton.className = "confirm"
        }
        confirmButton.addEventListener("click", () => {
            confirm()
            dialog.remove()
        })
        bottom.appendChild(confirmButton)
    }

    let checkValidateResult = () => {
        config.items.map(item => {
            validateResult[item.name] = {
                error: false,
                message: ""
            }
            let value = data[item.name] || "";
            if (item.required) {
                if (!value) {
                    validateResult[item.name] = {
                        error: true,
                        message: `${item.label}是必填项`
                    }
                    return
                }
            }
            if (item.regex) {
                if (!validateWithRegex(item.regex, value)) {
                    validateResult[item.name] = {
                        error: true,
                        message: `${item.regexerror}`
                    }
                }
            }
        })
        let errorMessage = Object.values(validateResult).reduce((a, b) => a + (b.error ? b.message + ";" : ""), "")
        return errorMessage
    }

    if (config.type === "form") {
        confirmButton.addEventListener("click", async () => {
            let checkResult = checkValidateResult()
            if (!checkResult) {
                formerror.style.display = "none"
                let result = await confirm(data)
                if (result === undefined || result === true) {
                    dialog.remove()
                }else if(typeof result === "string"){
                    formerror.innerText = result
                    formerror.style.display = "block"
                }
            } else {
                formerror.innerText = checkResult
                formerror.style.display = "block"
            }
        })
        bottom.appendChild(confirmButton)
    }

    bottom.appendChild(cancelButton)
    innerDialog.appendChild(bottom)

    dialog.appendChild(innerDialog)

    document.body.appendChild(dialog)

}

function alert(message) {
    alertDialog({ title: "提示", type: "message", text: message })
}

function confirm(message, confirm) {
    alertDialog({ title: "请确认", type: "confirm", text: message }, confirm)
}

function warn(message, confirm) {
    alertDialog({ title: "警告", type: "confirm", subtype: "danger", text: message }, confirm)
}

function toast(message){

    let toastbox = document.getElementById("toastbox")
    let toast = document.createElement("div")
    toast.innerText = message

    toastbox.appendChild(toast)
    setTimeout(()=>{
        toast.remove()
    }, 5000)
}

function modalDialog(message){
    let fullscreen = $(`<div style="position:fixed;left:0px;top:0px;right:0px;bottom:0px;background:rgba(0,0,0,0.6);display:flex;flex-direction: column;justify-content: center;align-items: center;font-size: 18px;color: white;z-index:10000">${message}</div>`)
    $("body").append(fullscreen)

    let closeDialog = () => {
        fullscreen.remove()
    }
    return closeDialog
}