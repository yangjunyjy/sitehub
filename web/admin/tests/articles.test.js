import test from 'node:test'
import assert from 'node:assert/strict'
import {articleListParams,normalizeArticle,articleMarkdown,articleFilename} from '../src/utils/articles.js'
test('article filters omit empty status, preserve draft zero and respect backend page limits',()=>{
  assert.deepEqual(articleListParams({title:'  ',status:'',categoryId:'',tagId:''},1,100),{page:1,pageSize:50})
  assert.deepEqual(articleListParams({title:' Vue ',status:'0',categoryId:2,tagId:3},2,20),{page:2,pageSize:20,title:'Vue',status:0,categoryID:'2',tagIDs:['3']})
})
test('article rows support nullable tags and export requires a complete detail response',()=>{
  assert.deepEqual(normalizeArticle({id:1,tags:null,status:0,is_top:true}).tags,[])
  assert.equal(normalizeArticle({id:1,tags:[null,{name:'Go'}],status:'1'}).tags.length,1)
  assert.throws(()=>articleMarkdown({id:1,title:'Summary only'}),/正文/)
  assert(articleMarkdown({title:'Content',content:'Complete body'}).includes('Complete body'))
  assert.notEqual(articleFilename({id:1,title:'Same name'}),articleFilename({id:2,title:'Same name'}))
})
