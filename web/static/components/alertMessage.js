customElements.define(
        'alert-message',
        class extends HTMLElement {
                constructor() {
                        super()
                }

                connectedCallback() {
                        this.classList.add('px-2', 'py-3', 'rounded', 'block', 'bold')

                        const variation = this.getAttribute('data-variation')

                        switch (variation) {
                                case 'success':
                                        this.classList.add('text-white', 'bg-green-800')
                                        break
                                case 'danger':
                                        this.classList.add('text-white', 'bg-red-900')
                                        break
                                case 'info':
                                        this.classList.add('text-white', 'bg-cyan-900')
                                        break
                        }
                }
        }
)
