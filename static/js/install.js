async function loadConfigSchema() {
    let configSchemaResp = await fetch("static/js/config.json")
    if (configSchemaResp.status === 200) {
        let configSchema = await configSchemaResp.json()
        return configSchema
    }
}

let container = $(".formcontainer")

let config = {}
window.onload = async () => {
    let schemas = await loadConfigSchema()
    for (let i = 0; i < schemas.length; i++) {
        let schema = schemas[i]
        generateConfigHtml(schema, newv => {
            config[schema.name] = newv
            clearDataKey(config, schema.name)
        })
    }

    $("#submitbtn").on("click", async () => {
        let check = validateData()
        if (!check) {
            alert("请填写所有必填项")
        } else {
            const response = await fetch("/", {
                method: "POST",
                body: JSON.stringify(config),
                headers: {
                    "Content-Type": "application/json"
                },
            });
            if (response.status === 200) {
                warn(`配置文件保存成功，它应该和程序在同一个目录，文件名是config.yaml。请稍等10秒然后刷新页面。
                    如果后续需要修改配置，请编辑或者删除config.yaml并重启服务器。
                    `)
            }else{
                let errtext = await response.text()
                warn(errtext)
            }
        }

    })
}

function validateData() {
    var visibleInputs = $('input:visible');
    let check = true
    for (let i = 0; i < visibleInputs.length; i++) {
        let input = visibleInputs[i]
        if (!input.checkValidity()) {
            check = false
            input.style.border = "1px solid red";
        }
    }
    return check
}

function generateConfigHtml(schema, valueChanged) {
    valueChanged = valueChanged || function () { }
    let div = $(`
        <div name="${schema.name}" class="schema">
            <div class="configtitle">
                <span class="title1">${schema.label}</span>
            </div>
            <div class="form"><div>
        </div>
        `)
    if (schema.desc) {
        div.find(".configtitle").append($(`<span class="desc">${schema.desc}</span>`))
    }
    div.find(".form").append(generateObjHtml(schema, v => {
        valueChanged(v)
    }))

    container.append(div)
}

function clearDataKey(data, key) {
    let v = data[key]
    if (v === undefined || v === null || v === "" || (typeof v === "object" && Object.keys(v).length === 0)) {
        delete data[key]
    }
}

function generateObjHtml(obj, valueChanged) {
    valueChanged = valueChanged || function () { }
    let data = {}

    let html = $(`<div class="border"></div>`)
    let props = obj.props
    for (let i = 0; i < props.length; i++) {
        let prop = props[i]
        let prophtml = $(`<div class="formitem">
            <div name="${prop.name}" class="formlabel"><span>${prop.label}</span></div>
            <div class="formvalue">
                <div class="value"></div>
            </div>
            </div>`)
        if (prop.type === "object") {
            let newprops_html = generateObjHtml(prop, v => {
                data[prop.name] = v
                clearDataKey(data, prop.name)
                valueChanged(data)
            })
            prophtml.find(".value").append(newprops_html)
        } else {
            let new_html = generateValueHtml(obj, prop, v => {
                data[prop.name] = v
                clearDataKey(data, prop.name)
                valueChanged(data)
            })
            prophtml.find(".value").append(new_html)
        }

        if (prop.desc) {
            prophtml.find(".formvalue").append($(`<div class="desc">${prop.desc}</div>`))
        }

        if (prop.required) {
            prophtml.find(".formlabel").append($(`<span style="color:red">*</span>`))
        }


        html.append(prophtml)
    }
    return html
}

let valueChangedListener = {}
let initValues = {}

function generateValueHtml(parent, prop, valueChanged) {
    valueChanged = valueChanged || function () { }
    let defaultvalue = prop.default || ""
    if (prop.type === "string") {
        let input = $(`<input class="input" name="${prop.name}" value="${defaultvalue}" />`)
        input.on("change", () => {
            let val = input.val()
            valueChanged(val)
            if (val) {
                input.css("border", "1px solid black")
            }
        })
        if (defaultvalue) {
            valueChanged(defaultvalue)
        }
        if (prop.required) {
            input.attr("required", "required")
        }

        return input
    }

    if (prop.type === "number" || prop.type === "integer") {
        let input = $(`<input class="input" name="${prop.name}" value="${defaultvalue}" type="number" style="width:100px" />`)
        input.on("change", () => {
            let val = input.val()
            if(prop.reqtype === "string"){
                valueChanged(val + "")
            }else{
                valueChanged(val)
            }

            
            if (val) {
                input.css("border", "1px solid black")
            }
        })
        if (defaultvalue) {
            if(prop.reqtype === "string"){
                valueChanged(defaultvalue + "")
            }else{
                valueChanged(defaultvalue)
            }
        }
        if (prop.required) {
            input.attr("required", "required")
        }
        return input
    }

    if (prop.type === "boolean") {
        let input = $(`<input class="checkbox" name="${prop.name}" type="checkbox" />`)
        input.on("change", () => {
            if(prop.reqtype === "string"){
                valueChanged(input.is(":checked") + "")
            }else{
                valueChanged(input.is(":checked"))
            }
            
        })
        if (defaultvalue) {
            input.attr("checked", "checked")
            if(prop.reqtype === "string"){
                valueChanged(defaultvalue + "")
            }else{
                valueChanged(defaultvalue)
            }
            
        }
        if (prop.required) {
            input.attr("required", "required")
        }
        return input
    }

    if (prop.type === "option") {
        let optionDiv = $(`<select name="${prop.name}" class="options"></select>`)

        if (parent.required !== true) {
            optionDiv.append($(`<option value="" desc="">不设置</option>`))
        }

        let options = prop.options
        for (let i = 0; i < options.length; i++) {
            optionDiv.append($(`<option value="${options[i].value}" desc="${options[i].desc || ""}">${options[i].label}</option>`))
        }

        let listenerkey = parent.name + "_" + prop.name
        optionDiv.on("input", () => {
            let val = optionDiv.val()
            valueChanged(val)
            let desc = optionDiv.find(":selected").attr("desc")
            let listeners = valueChangedListener[listenerkey]
            if (listeners && listeners.length > 0) {
                for (let i = 0; i < listeners.length; i++) {
                    listeners[i](val,desc)
                }
            }
        })

        initValues[listenerkey] = {
            value: optionDiv.val(),
            desc: optionDiv.find("option").eq(0).attr("desc")
        }

        if (parent.required === true) {
            valueChanged(optionDiv.val())
        }

        return optionDiv
    }

    if (prop.type === "ref") {
        let listenerkey = parent.name + "_" + prop.ref
        if (valueChangedListener[listenerkey] === undefined) {
            valueChangedListener[listenerkey] = []
        }

        let mem = {}

        let cprops = prop.props
        let html = $(`<div>
            <div class="optiondesc"></div>
            </div>`)
        for (let i = 0; i < cprops.length; i++) {
            let cprop = cprops[i]
            let c = $(`<div class="c" id="${cprop.value}" style="display:none"></div>`)
            c.append(generateObjHtml(cprop, v => {
                mem[cprop.value] = v
                valueChanged(v)
            }))
            html.append(c)
        }

        let listener = (newdata, desc) => {
            desc = desc || ""
            html.find(".optiondesc").text(desc)
            html.find(".c").css("display", "none")
            if (newdata !== "") {
                html.find(`#${newdata}`).css("display", "block")
            }
            valueChanged(mem[newdata])
        }
        valueChangedListener[listenerkey].push(listener)
        listener(initValues[listenerkey].value, initValues[listenerkey].desc)

        return html
    }
}