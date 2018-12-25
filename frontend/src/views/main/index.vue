<template>
  <split-pane :default-percent="20" split="vertical" class="components-container">
    <template slot="paneL" >
      <KeysList/>
    </template>
    <template slot="paneR">
      <Edit/>
    </template>
  </split-pane>
</template>

<script>
import { mapGetters } from 'vuex'
import splitPane from 'vue-splitpane'
import KeysList from './components/KeysList'
import Edit from './components/Edit'

export default {
  name: 'Main',
  components: { splitPane, KeysList, Edit },
  data() {
    return {
      loading: false,
      keys: []
    }
  },
  computed: {
    ...mapGetters([
      'name',
      'roles'
    ])
  },
  watch: {
    $route: {
      handler: function(route) {
        // console.log(this.$route.query)
        // console.log(this.$route.params)
      },
      immediate: true
    }
  },
  methods: {
  }
}
</script>

<style rel="stylesheet/scss" lang="scss" scoped>
  .components-container {
    position: relative;
    height: 95vh;
  }

  .left-container {
    height: 100%;
  }

  .right-container {
    height: 200px;
  }

  .top-container {
    width: 100%;
    height: 100%;
  }

  .bottom-container {
    width: 100%;
    height: 100%;
  }
</style>
