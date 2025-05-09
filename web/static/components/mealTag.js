customElements.define(
    'meal-tag',
    class extends HTMLElement {
        constructor() {
            super()
        }

        template(existingTag) {
            const node = document.createElement('div')

            node.classList.add('flex', 'items-center', 'gap-4', 'wrap')

            const body = `
            <div class="flex flex-col gap-2">
                <label class="text-sm text-slate-300" for="tagName">name</label>
                <request-typeahead
                    data-name="tagName"
                    data-url="http://localhost:8000/api/tags"
                />
                <input type="hidden" name="tagId" value="0" />
            </div>
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

            template.querySelector('[data-name="tagName"]').dataset.value = this.getAttribute('data-name')

            if (existingTag) {
                template.querySelector('[name="tagId"]').value = id
            }

            template.querySelector('.remove-tag').onclick = () => {
                document.querySelector('meal-tag[data-index="' + index + '"]').remove()
            }

            this.appendChild(template)
        }
    }
)
