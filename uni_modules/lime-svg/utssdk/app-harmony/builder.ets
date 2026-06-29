// import { Color } from '@ohos.arkui';
@Builder
export function buildImage(params: ESObject) {
	if(params.color) {
		Image(params.src)
			.width('100%')
			.height('100%')
			.objectFit(ImageFit.Contain)
		    // .fillColor(Color.Blue) 
		    .fillColor(params.color) 
		    // colorFilter只对位图生效
			.colorFilter(params.colorFilter)
			// .colorFilter(
   //          [1, 0, 0, 0, 1,
   //           0, 1, 0, 0, 1,
   //           0, 0, 1, 0, 0,
   //           0, 0, 0, 1, 0])
			.onComplete((event)=>{
				params.onComplete(event)
			})
			.onError((error) =>{
				params.onError(error.message)
              })
	} else {
		Image(params.src)
			.width('100%')
			.height('100%')
			.objectFit(ImageFit.Contain)
			.onComplete((event)=>{
				params.onComplete(event)
			})
			.onError((error) =>{
				params.onError(error.message)
            })
	}
}
