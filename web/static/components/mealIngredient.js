customElements.define(
        'meal-ingredient',
        class extends HTMLElement {
                constructor() {
                        super()
                }

                template() {
                        const node = document.createElement('div')

                        node.classList.add('flex', 'items-center', 'gap-4', 'wrap')

                        const body = `
                <div class="flex flex-col gap-2">
                    <label class="text-sm text-slate-300" for="ingredientName">name</label>
                        <request-typeahead
                            data-name="ingredientName"
                            data-url="http://localhost:8000/api/ingredients"
                        />
                        <input type="hidden" name="ingredientId" value="0" />
                </div>
            <input name="ingredientId" value="0" type="hidden" />
            <div class="flex flex-col gap-2">
                <label class="text-sm text-slate-300" for="ingredientQuantity">quantity</label>
                <input name="ingredientQuantity" class="quantity | py-2 px-1.5 rounded bg-zinc-700 ring-1 ring-zinc-400 w-[250px]" />
            </div>
            <div class="flex flex-col gap-2">
                <label class="text-sm text-slate-300" for="ingredientUnit">unit</label>
                <input name="ingredientUnit" class="unit | py-2 px-1.5 rounded bg-zinc-700 ring-1 ring-zinc-400 w-[250px]" />
            </div>
            <button class="remove-ingredient | py-2 px-4 rounded-md text-red-400 hover:text-slate-300 hover:bg-red-900 transition-colors border border-red-400 hover:border-red-900 self-end">Remove</button>
            `

                        node.innerHTML = body

                        return node
                }

                connectedCallback() {
                        const id = this.getAttribute('data-id')

                        const existingIngredient = id !== null && id > 0
                        const template = this.template()

                        const index = this.getAttribute('data-index')

                        if (this.hasAttribute('data-name')) {
                                template.querySelector('[data-name="ingredientName"]').dataset.value = this.getAttribute('data-name')
                        }
                        template.querySelector('.quantity').value = this.getAttribute('data-quantity')
                        template.querySelector('.unit').value = this.getAttribute('data-unit')

                        if (existingIngredient) {
                                template.querySelector('[name="ingredientId"]').value = id
                        }

                        template.querySelector('.remove-ingredient').onclick = () => {
                                document.querySelector('meal-ingredient[data-index="' + index + '"]').remove()
                        }

                        this.appendChild(template)
                }
        }
)
