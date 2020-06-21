<template>
  <div>
    <h3 id="tituloListaInterno"><b>List of all checked pages</b></h3>
    <div v-if="propList.length > 0">
      <ul class="list-group list-group-flush">
        <li class="list-group-item" v-for="element in propList" :key="element.name">{{element}}</li>
      </ul>
    </div>
    <div v-else>
      <p>No websites have been checked yet.</p>
    </div>
  </div>
</template>

<script>
export default {
  name: "MyList",
  props: { propList: { type: Array } }
};
</script>

<style scoped>
h3 {
  margin: 40px 0 0;
}
ul {
  list-style-type: none;
  padding: 0;
}
li {
  display: inline-block;
  margin: 0 10px;
}
a {
  color: #42b983;
}

#tituloListaInterno{
  margin-bottom: 1em;
}

</style>