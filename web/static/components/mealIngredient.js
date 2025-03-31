customElements.define(
    'meal-ingredient',
    class extends HTMLElement {
        constructor() {
            super()
        }

        template() {
            const node = document.createElement('div')

            node.classList.add('flex', 'items-center', 'gap-4', 'wrap')
            node.innerHTML = `
            <div class="flex flex-col gap-2">
                <label class="text-sm text-slate-300" for="notes">name</label>
                <input name="ingredientName" class="name | py-2 px-1.5 rounded bg-zinc-700 ring-1 ring-zinc-400 w-[250px] disabled:bg-zinc-500" />
                <input name="ingredientId" value="0" type="hidden" />
            </div>
            <div class="flex flex-col gap-2">
                <label class="text-sm text-slate-300" for="notes">quantity</label>
                <input name="ingredientQuantity" class="quantity | py-2 px-1.5 rounded bg-zinc-700 ring-1 ring-zinc-400 w-[250px]" />
            </div>
            <div class="flex flex-col gap-2">
                <label class="text-sm text-slate-300" for="notes">unit</label>
                <input name="ingredientUnit" class="unit | py-2 px-1.5 rounded bg-zinc-700 ring-1 ring-zinc-400 w-[250px]" />
            </div>
            <label class="flex gap-2 self-end">
                <input type="radio" name="isMain" />
                <span>Main Ingredient</span>
            </label>
            <button class="remove-ingredient | py-2 px-1.5 rounded-md bg-cyan-800 hover:bg-cyan-900 transition-colors w-auto self-end">remove</button>
            `

            return node
        }

        connectedCallback() {
            const template = this.template()

            const index = this.getAttribute('data-index')

            const nameEl = template.querySelector('.name')

            nameEl.value = this.getAttribute('data-name')
            template.querySelector('.quantity').value = this.getAttribute('data-quantity')
            template.querySelector('.unit').value = this.getAttribute('data-unit')
            template.querySelector('[name="isMain"]').value = index

            const id = this.getAttribute('data-id')

            if (id !== null && id > 0) {
                template.querySelector('[name="ingredientId"]').value = id
                nameEl.disabled = true
            }

            template.querySelector('.remove-ingredient').onclick = () => {
                document.querySelector('meal-ingredient[data-index="' + index + '"]').remove()
            }

            this.appendChild(template)
        }
    }
)
