customElements.define(
    'typeahead-result',
    class extends HTMLElement {
        constructor() {
            super()
        }

        template() {
            const node = document.createElement('li')

            node.classList.add('cursor-pointer', 'px-2', 'py-3', 'hover:bg-slate-600', 'transition-colors')

            return node
        }

        connectedCallback() {
            const template = this.template()
            const id = this.getAttribute('data-id')
            const label = this.getAttribute('data-label')

            template.innerText = label
            template.onclick = () => {
                const event = new CustomEvent('selected', {
                    detail: { id, label },
                    bubbles: true,
                })

                this.dispatchEvent(event)
            }

            this.appendChild(template)
        }
    }
)

customElements.define(
    'request-typeahead',
    class extends HTMLElement {
        constructor() {
            super()

            this.requestController = null
            this.requestUrl = null
        }

        template() {
            const node = document.createElement('div')

            node.classList.add('flex', 'flex-col')

            node.innerHTML = `
            <input 
                class="py-2 px-1.5 rounded bg-zinc-700 ring-1 ring-zinc-400 w-[250px]"
            />
            <div class="relative">
                <ul class="results | hidden absolute top-[5px] bg-slate-500 left-0 right-0 rounded"></ul>
            </div>
            `

            return node
        }

        connectedCallback() {
            const template = this.template()

            const input = template.querySelector('input')

            this.requestUrl = this.getAttribute('data-url')

            input.addEventListener('keyup', (e) => this.search(e.target.value))

            input.placeholder = this.getAttribute('data-placeholder')

            this.appendChild(template)

            this.addEventListener('selected', () => this.reset())
        }

        reset() {
                const resultsContainer = this.querySelector('.results')
                const input = this.querySelector('input')

                input.value = ''

                resultsContainer.classList.add('hidden')
                resultsContainer.innerHTML = ''
        }

        async search(query) {
            if (this.requestController !== null) {
                this.requestController.abort()
            }

            this.requestController = new AbortController

            const response = await fetch(this.requestUrl + '?query=' + query, {
                signal: this.requestController.signal
            })

            const results = await response.json()

            const resultsContainer = this.querySelector('.results')

            resultsContainer.innerHTML = ''

            results.forEach(result => {
                const node = document.createElement('typeahead-result')

                node.dataset.id = result.id
                node.dataset.label = result.name

                resultsContainer.append(node)
            })

            resultsContainer.classList.remove('hidden')
        }
    }
)
