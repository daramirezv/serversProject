<template>
  <div>
    <div v-if="propMessage === '' && propSsl === ''">
      <h4>No query has been made so far</h4>
    </div>
    <div v-else-if="propSsl === ''">
      <h4>{{propMessage}}</h4>
    </div>
    <div v-else>
      <h3 class="titulosespecificos">General website information</h3>
      <table class="table table-striped">
        <thead>
          <tr>
            <th scope="col">Field</th>
            <th scope="col">Value</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td scope="row">servers_changed</td>
            <td>{{propServersChanged}}</td>
          </tr>
          <tr>
            <td scope="row">ssl_grade</td>
            <td>{{propSsl}}</td>
          </tr>
          <tr>
            <td scope="row">previous_ssl_grade</td>
            <td>{{propPrevSsl}}</td>
          </tr>
          <tr>
            <td scope="row">logo</td>
            <td>{{propLogo}}</td>
          </tr>
          <tr>
            <td scope="row">title</td>
            <td>{{propTitle}}</td>
          </tr>
          <tr>
            <td scope="row">is_down</td>
            <td>{{propDown}}</td>
          </tr>
        </tbody>
      </table>
      <h3 class="titulosespecificos">Specific servers information</h3>
      <table class="table table-striped ultimaTabla">
        <thead>
          <tr>
            <th scope="col">#</th>
            <th scope="col">address</th>
            <th scope="col">ssl_grade</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(element, index) in propTableIP" :key="index">
            <th scope="row">{{index+1}}</th>
            <td>{{element}}</td>
            <td>{{propTableGrade[index]}}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script>
export default {
  name: "MyTable",
  props: {
    propTableIP: { type: Array },
    propTableGrade: { type: Array },
    propServersChanged: { type: Boolean },
    propSsl: { type: String },
    propPrevSsl: { type: String },
    propLogo: { type: String },
    propTitle: { type: String },
    propDown: { type: Boolean },
    propMessage: { type: String }
  }
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

.titulosespecificos{
  margin-bottom: 1em;
}

.ultimaTabla {
  margin-bottom: 5em;
}

</style>