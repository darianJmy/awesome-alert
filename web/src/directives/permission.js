import { useUserStore } from '../stores/user'

export default {
  mounted(el, binding) {
    const { value } = binding
    if (value && value.length > 0) {
      const store = useUserStore()
      if (!value.some((code) => store.permissions.includes(code))) {
        el.parentNode?.removeChild(el)
      }
    }
  },
}
