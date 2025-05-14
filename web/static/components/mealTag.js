customElements.define(
        'meal-tag',
        class extends HTMLElement {
                constructor() {
                        super()
                }

                template() {
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
                        <button class="remove-tag | py-2 px-4 rounded-md text-red-400 hover:text-slate-300 hover:bg-red-900 transition-colors border border-red-400 hover:border-red-900 self-end">Remove</button>
                        `

                        node.innerHTML = body

                        return node
                }

                connectedCallback() {
                        const id = this.getAttribute('data-id')

                        const existingTag = id !== null && id > 0
                        const template = this.template()

                        const index = this.getAttribute('data-index')

                        if (this.hasAttribute('data-name')) {
                                template.querySelector('[data-name="tagName"]').dataset.value = this.getAttribute('data-name')
                        }

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
