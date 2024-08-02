function isEmpty(text){
    if(text){
        text = text.replaceAll("\n","").replaceAll(" ","").replaceAll(" ","")
        return text === ""
    }
    return true
}
class CardBox {
    constructor() {
        this.init()
        this.reset()
        this.content = ""
    }

    async init() {
        let buttons = $(".cardbuttons")
        let cardedit = $(".cardedit")
        let configResp = await fetch("/api/v1/post/card/config")
        if (configResp.status === 200) {
            let configmap = (await configResp.json()).data
            this.configmap = configmap
            Object.keys(configmap).forEach(key => {
                let name = key
                let config = configmap[name]
                let displayname = config.displayname
                let button = $(`<button>${displayname}</button>`)
                button.on("click", () => {
                    this.model = {}
                    this.name = name
                    cardedit.html("")
                    this.dataChanged()
                    buttons.find("button").removeClass("click")
                    button.addClass("click")
                    let props = config.props
                    for (let i = 0; i < props.length; i++) {
                        let prop = props[i]
                        let input;
                        if (prop.type === 'string') {
                            if (prop.long) {
                                input = $(`<textarea class="v" placeholder="请输入${prop.displayname}" ></textarea>`)
                            } else {
                                input = $(`<input type="text" class="v" name="${prop.name}" placeholder="请输入${prop.displayname}" />`)
                            }
                            input.on("blur", () => {
                                this.dataChanged(prop.name, input.val())
                            })

                        }
                        if (prop.type === 'number') {
                            input = $(`<input type="number" class="v" name="${prop.name}"  placeholder="请输入${prop.displayname}"/>`)
                            input.on("blur", () => {
                                this.dataChanged(prop.name, input.val())
                            })

                        }
                        if (prop.type === 'date') {
                            input = $(`<input type="date" class="v" name="${prop.name}"  placeholder="请输入${prop.displayname}"/>`)
                            input.on("blur", () => {
                                this.dataChanged(prop.name, input.val())
                            })

                        }
                        if (prop.type === 'time') {
                            input = $(`<input type="time" class="v" name="${prop.name}"  placeholder="请输入${prop.displayname}"/>`)
                            input.on("blur", () => {
                                this.dataChanged(prop.name, input.val())
                            })

                        }
                        if (prop.type === 'datetime') {
                            input = $(`<input type="datetime-local" class="v" name="${prop.name}"  placeholder="请输入${prop.displayname}"/>`)
                            input.on("blur", () => {
                                this.dataChanged(prop.name, input.val())
                            })

                        }
                        if (prop.type === 'image') {
                            let width = 100;
                            let height = 100;
                            if (prop.aspectRadio) {
                                let xy = prop.aspectRadio.split(":")
                                let x = parseInt(xy[0])
                                let y = parseInt(xy[1])
                                width = Math.sqrt(10000 / (x * y)) * x;
                                height = 10000 / width
                            }
                            input = $(`<div class="imageinput" style="width:${width}px;height:${height}px"></div>`)
                            new Uploader(input, {
                                types: ["image"],
                                allowSoftDelete: true,
                                deleteCallback: () => {
                                    this.dataChanged(prop.name, undefined)
                                },
                                callback: uploadResp => {
                                    this.dataChanged(prop.name, {
                                        url: uploadResp.url
                                    })
                                },
                            })
                        }
                        if (prop.type === 'link') {
                            input = $(`<div>
                                    <input type="text" class="v1" name="${prop.name}_name" placeholder="请输入${prop.displayname}名称" /><br>
                                    <input type="text" class="v2" name="${prop.name}_link" placeholder="请输入${prop.displayname}地址" style="border-top:1px solid #E4E7ED;margin-top:5px;padding-top:5px" />
                               </div>`)

                            let v1 = input.find(".v1")
                            let v2 = input.find(".v2")
                            v1.on("change", () => {
                                this.dataChanged(prop.name, {
                                    title: v1.val(),
                                    url: v2.val()
                                })
                            })
                            v2.on("change", () => {
                                this.dataChanged(prop.name, {
                                    title: v1.val(),
                                    url: v2.val()
                                })
                            })
                        }
                        if (prop.type === 'option') {
                            let options = ""
                            for (let j = 0; j < prop.options.length; j++) {
                                if (j === 0) {
                                    options += `<option value="${prop.options[j].value}" selected="selected">${prop.options[j].label}</option>`
                                } else {
                                    options += `<option value="${prop.options[j].value}" >${prop.options[j].label}</option>`
                                }

                            }
                            input = $(`<select class="v">${options}</select>`)
                            input.on("change", () => {
                                this.dataChanged(prop.name, {
                                    value: input.val()
                                })
                            })
                        }
                        if (prop.type === 'boolean') {
                            input = $(`<input type="checkbox" class="v" name="${prop.name}"/>`)
                            input.on("change", () => {
                                this.dataChanged(prop.name, input.is(':checked'))
                            })
                        }


                        if(prop.type === "array"){
                            let min = prop.min
                            let max = prop.max
                            input = $(`<div class="array"><div class="itemsinputs"></div><a href="javascript:void(0)" style="display:inline-block" class="addBtn">增加一项</a></div>`)
                            let itemsinputs = input.find(".itemsinputs")
                            let addBtn = input.find(".addBtn")
                            let dataChanged = ()=>{
                                let values = []
                                let inputs = itemsinputs.find("input")
                                for(let i = 0; i < inputs.length; i++){
                                    values.push(inputs[i].value)
                                }
                                this.dataChanged(prop.name, values)

                            }
                            addBtn.on("click",()=>{
                                let iteminput = $(`<div><input type="text" style="margin-bottom:5px"  placeholder="请输入内容"/><a class="deletebtn" href="javascript:void(0)" style="display:none">删除</a></div>`)
                                let deletebtn = iteminput.find(".deletebtn")
                                deletebtn.on("click", ()=>{
                                    if(itemsinputs.find("input").length > min){
                                        iteminput.remove()
                                        addBtn.css("display", "inline-block")
                                        if(itemsinputs.find("input").length <= min){
                                            itemsinputs.find(".deletebtn").css("display", "none")
                                        }
                                        dataChanged()
                                    }
                                })
                                let input = iteminput.find("input")
                                input.on("blur", ()=>{
                                    dataChanged()
                                })
                                itemsinputs.append(iteminput)
                                if(itemsinputs.find("input").length >= max){
                                    addBtn.css("display", "none")
                                }

                                if(itemsinputs.find("input").length > min){
                                    itemsinputs.find(".deletebtn").css("display", "inline-block")
                                }

                            })
                           
                            for(let i = 0; i < min; i++){
                                addBtn.click();
                            }
                            
                        }

                        let inputbox = $(`<div class="item">
                            <div class="label">${prop.displayname}</div>
                            <div class="value">
                                
                            </div>
                        </div>`)

                        inputbox.find(".value").append(input)
                        if (prop.required) {
                            inputbox.find(".label").append(`<span style="color:red">&nbsp;*</span>`)
                        }

                        cardedit.append(inputbox)
                    }
                })
                buttons.append(button)
            })
        }

        this.buttons = buttons
    }

    async dataChanged(name, value) {
        if (name) {
            this.model[name] = value
        }

        let previewResp = await fetch(`/api/v1/post/card/preview?name=${this.name}&model=${JSON.stringify(this.model)}&content=${this.content}`)
        if (previewResp.status === 200) {
            let html = (await previewResp.json()).data
            let cardhtml = $(html)
            let skeletonEles = cardhtml.find("[skeleton]")
            for(let i = 0; i < skeletonEles.length; i++){
                let skeletonEle = $(skeletonEles[i])
                let skeletonAttr = skeletonEle.attr("skeleton")
                if(!skeletonAttr){
                    skeletonAttr = "text"
                }
                let skeletonArr = skeletonAttr.split(";")
                let type = skeletonArr[0]
                if(type === "text"){
                    if(isEmpty(skeletonEle.text())){
                        // 填充骨架
                        let width = ["100%"]
                        let line = 1
                        for(let j = 1; j < skeletonArr.length; j++){
                            if(skeletonArr[j].indexOf("=") !== -1){
                                let kv = skeletonArr[j].split("=")
                                if(kv[0] === "width"){
                                    width = kv[1].split(",")
                                }
                                if(kv[0] === "line"){
                                    line = parseInt(kv[1])
                                }
                            }
                        }

                        for(let j = 0; j < line; j++){
                            let curwidth = width.length > j ? width[j] : width[0];
                            if(j !== 0){
                                skeletonEle.append("<br>");
                            }
                            skeletonEle.append(`<span style="width:${curwidth};display:inline-block;background: #f2f2f2;">&nbsp;</span>`)  
                        }
                        
                    }
                }
                if(type === "image"){
                    let src = skeletonEle.attr("src")
                    if(src === ""){
                        skeletonEle.attr("src","/static/images/empty.jpg")
                    }
                }
                if(type === "button"){
                    if(isEmpty(skeletonEle.text())){
                        let text = "示例按钮"
                        for(let j = 1; j < skeletonArr.length; j++){
                            if(skeletonArr[j].indexOf("=") !== -1){
                                let kv = skeletonArr[j].split("=")
                                if(kv[0] === "text"){
                                    text = kv[1]
                                }
                            }
                        }
                        skeletonEle.text(text)
                    }
                }
                if(type === "list"){
                    if(isEmpty(skeletonEle.text())){
                         // 填充骨架
                         let width = ["100%"]
                         let line = 1
                         for(let j = 1; j < skeletonArr.length; j++){
                             if(skeletonArr[j].indexOf("=") !== -1){
                                 let kv = skeletonArr[j].split("=")
                                 if(kv[0] === "width"){
                                     width = kv[1].split(",")
                                 }
                                 if(kv[0] === "line"){
                                     line = parseInt(kv[1])
                                 }
                             }
                         }
                         for(let j = 0; j < line; j++){
                            let curwidth = width.length > j ? width[j] : width[0];
                            skeletonEle.append(`<div style="width:${curwidth};display:flex;margin-top:10px"><span style="width:20px;display:inline-block;background: #f2f2f2;margin-right:5px;">&nbsp;</span> <span style="flex:1;display:inline-block;background: #f2f2f2;">&nbsp;</span></div>`)  
                        }
                    }
                }

            }
            $(".cardpreview").html(cardhtml)
        }
    }

    checkData() {
        if (this.model) {
            let props = this.configmap[this.name].props
            for (let i = 0; i < props.length; i++) {
                if (props[i].required) {
                    let value = this.model[props[i].name]
                    if (!value) {
                        return false
                    }
                    if (props[i].type === 'image' && !value.url) {
                        return false;
                    }
                    if (props[i].type === 'link' && (!value.url || !value.title)) {
                        return false;
                    }
                    if (props[i].type === 'option' && (!value.value)) {
                        return false;
                    }
                    if(props[i].type === 'list'){
                        if(value.length < props[i].min || value.length > props[i].max){
                            return false;
                        }
                    }

                }
            }
            return true
        }
        return false
    }

    reset() {
        if (this.buttons) {
            this.buttons.find("button")[0].click()
        }

    }
    clear(){
        $(".cardedit").html("")
        this.model = undefined
    }

    setContent(content){
        this.content = content
        if(this.model){
            this.dataChanged()
        }
    }

    getData() {
        if (this.model && Object.keys(this.model).length > 0) {
            return {
                name: this.name,
                model: this.model
            }
        }
        return undefined
    }
}
