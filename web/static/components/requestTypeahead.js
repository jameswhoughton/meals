customElements.define(
        'typeahead-result',
        class extends HTMLElement {
                constructor() {
                        super()
                }

                template() {
                        const node = document.createElement('li')
                        node.role = 'tab'
                        node.tabIndex = 0

                        const classes = [
                                'cursor-pointer',
                                'px-2',
                                'py-3',
                                'hover:bg-slate-600',
                                'focus:bg-slate-600',
                                'transition-colors',
                        ]

                        node.classList.add(...classes)

                        return node
                }

                connectedCallback() {
                        const template = this.template()
                        const id = this.getAttribute('data-id')
                        const label = this.getAttribute('data-label')

                        template.innerText = label
                        const selectedFn = () => {
                                const event = new CustomEvent('selected', {
                                        detail: { id, label },
                                        bubbles: true,
                                })

                                this.dispatchEvent(event)
                        }

                        template.addEventListener('click', () => selectedFn())
                        template.addEventListener('keyup', (e) => {
                                if (e.keyCode === 13) {
                                        selectedFn()
                                }
                        })

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
                        this.debounceTimeout = null
                }

                template() {
                        const node = document.createElement('div')

                        node.classList.add('flex', 'flex-col')

                        node.innerHTML = `
            <input 
                    class="py-2 px-1.5 rounded bg-zinc-700 ring-1 ring-zinc-400"
                aria-autocomplete="list"
                aria-expanded="false"
                autocomplete="off"
            />
            <div class="relative">
                <ul class="results | hidden absolute top-[5px] bg-slate-500 left-0 right-0 rounded" role="tablist"></ul>
            </div>
            `

                        return node
                }

                connectedCallback() {
                        const id = this.id

                        if (id === undefined) {
                                throw new Exception('id attribute required for request-typeahead')
                        }

                        const template = this.template()

                        const input = template.querySelector('input')
                        const results = template.querySelector('.results')

                        results.id = id + '-results'

                        // Set aria attributes
                        input.setAttribute('aria-controls', results.id)

                        this.requestUrl = this.getAttribute('data-url')

                        // Hide results if the escape key is pressed
                        this.addEventListener('keyup', (e) => {
                                if (e.keyCode === 27) {
                                        this.hide()
                                }
                        })

                        input.addEventListener('keyup', (e) => this.debounceSearch(e.target.value))

                        input.placeholder = this.getAttribute('data-placeholder') ?? ''

                        input.value = this.getAttribute('data-value')
                        input.name = this.getAttribute('data-name')

                        this.appendChild(template)

                        this.addEventListener('selected', (e) => {
                                input.value = e.detail.id

                                this.hide()
                        })
                }

                reset() {
                        const resultsContainer = this.querySelector('.results')
                        const input = this.querySelector('input')

                        input.value = ''

                        this.hide()
                        resultsContainer.innerHTML = ''
                }

                show() {
                        this.querySelector('.results').classList.remove('hidden')
                        this.querySelector('input').setAttribute('aria-expanded', true)
                }

                hide() {
                        this.querySelector('.results').classList.add('hidden')
                        this.querySelector('input').setAttribute('aria-expanded', false)
                }

                debounceSearch(query) {
                        if (this.debounceTimeout !== null) {
                                clearTimeout(this.debounceTimeout)
                        }

                        this.debounceTimeout = setTimeout(() => this.search(query), 300)
                }

                async search(query) {

                        if (query === '') {
                                return
                        }

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

                                node.dataset.id = result
                                node.dataset.label = result

                                resultsContainer.append(node)
                        })

                        this.show()
                }
        }
)
