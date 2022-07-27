package hardloop

import (
	"reflect"
	"testing"
	"time"
)

func TestSpecSchedule_Prev(t *testing.T) {
	type fields struct {
		CronSpec string
	}
	type args struct {
		t time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    time.Time
		wantErr bool
	}{
		{
			name: "test [0 7 * * *]",
			fields: fields{
				CronSpec: "0 7 * * *",
			},
			args: args{
				// 10:00:00
				t: time.Date(2020, time.January, 1, 10, 0, 0, 0, time.UTC),
			},
			// 07:00:00
			want:    time.Date(2020, time.January, 1, 7, 0, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name: "test [0 17 * * 1,2,3,4,5]",
			fields: fields{
				CronSpec: "0 17 * * 1,2,3,4,5",
			},
			args: args{
				t: time.Date(2020, time.January, 5, 10, 0, 0, 0, time.UTC),
			},
			want:    time.Date(2020, time.January, 3, 17, 0, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name: "test [35 17 * * 1,2,3,4,5]",
			fields: fields{
				CronSpec: "35 17 * * 1,2,3,4,5",
			},
			args: args{
				t: time.Date(2020, time.January, 5, 10, 20, 0, 0, time.UTC),
			},
			want:    time.Date(2020, time.January, 3, 17, 35, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name: "test [35 17 * * 1,2,3,4,5]",
			fields: fields{
				CronSpec: "35 17 * * 1,2,3,4,5",
			},
			args: args{
				t: time.Date(2020, time.January, 5, 10, 50, 0, 0, time.UTC),
			},
			want:    time.Date(2020, time.January, 3, 17, 35, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name: "test [0 17 23 5 *]",
			fields: fields{
				CronSpec: "0 17 23 5 *",
			},
			args: args{
				t: time.Date(2020, time.January, 5, 10, 50, 0, 0, time.UTC),
			},
			want:    time.Date(2019, time.May, 23, 17, 0, 0, 0, time.UTC),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := ParseStandard(tt.fields.CronSpec)
			if err != nil != tt.wantErr {
				t.Errorf("ParseStandard() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if got := s.Prev(tt.args.t); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SpecSchedule.Prev() = %v, want %v", got, tt.want)
			}
		})
	}
}
