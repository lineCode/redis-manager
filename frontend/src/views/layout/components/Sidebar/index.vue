<template>
  <el-scrollbar wrap-class="scrollbar-wrapper">
    <el-menu
      :show-timeout="200"
      :default-active="$route.path"
      :collapse="isCollapse"
      mode="vertical"
      background-color="#304156"
      text-color="#bfcbd9"
      active-text-color="#409EFF"
    >
      <app-link v-for="server in serverList" :key="server.id" :to="resolvePath(server.id)">
        <el-menu-item :index="resolvePath(server.id)" class="submenu-title-noDropdown">
          <item :title="server.name" icon="nested"/>
        </el-menu-item>
      </app-link>
    </el-menu>
  </el-scrollbar>
</template>

<script>
import path from 'path'
import { mapGetters } from 'vuex'
import { getServerList } from '@/api/server'
import Item from './Item'
import AppLink from './Link'

export default {
  components: { AppLink, Item },
  data() {
    return {
      serverList: []
    }
  },
  computed: {
    ...mapGetters([
      'sidebar'
    ]),
    isCollapse() {
      return !this.sidebar.opened
    }
  },
  created() {
    getServerList().then((response) => {
      this.serverList = response.data
    })
  },
  methods: {
    resolvePath(serverID) {
      return path.resolve(this.basePath, '/server/' + serverID)
    }
  }
}
</script>
