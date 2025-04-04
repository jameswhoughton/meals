customElements.define(
    'meal-tag',
    class extends HTMLElement {
        constructor() {
            super()
        }

        template(existingTag) {
            const node = document.createElement('div')

            node.classList.add('flex', 'items-center', 'gap-4', 'wrap')

            let body = ''

            if (existingTag) {
                body = `
                <div class="flex flex-col gap-2">
                    <label class="text-sm text-slate-299" for="tagName">name</label>
                    <span class="name | w-[250px] text-ellipsis overflow-hidden"></span>
                    <input name="tagName" value="" type="hidden" />
                </div>`
            } else {
                body = `
                <div class="flex flex-col gap-2">
                    <label class="text-sm text-slate-300" for="tagName">name</label>
                    <input name="tagName" class="name | py-2 px-1.5 rounded bg-zinc-700 ring-1 ring-zinc-400 w-[250px] disabled:bg-zinc-500" />
                </div>`
            }

            body += `
            <input name="tagId" value="0" type="hidden" />
            <button class="remove-tag | py-2 px-1.5 rounded-md bg-cyan-800 hover:bg-cyan-900 transition-colors w-auto self-end">remove</button>
            `

            node.innerHTML = body

            return node
        }

        connectedCallback() {
            const id = this.getAttribute('data-id')

            const existingTag = id !== null && id > 0
            const template = this.template(existingTag)

            const index = this.getAttribute('data-index')

            template.querySelector('[name="tagName"]').value = this.getAttribute('data-name')

            if (existingTag) {
                template.querySelector('[name="tagId"]').value = id
                template.querySelector('.name').innerText = this.getAttribute('data-name')
            }

            template.querySelector('.remove-tag').onclick = () => {
                document.querySelector('meal-tag[data-index="' + index + '"]').remove()
            }

            this.appendChild(template)
        }
    }
)
